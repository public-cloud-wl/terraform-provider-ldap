package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLDAPGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceLDAPGroupCreate,
		Read:   resourceLDAPGroupRead,
		Update: resourceLDAPGroupUpdate,
		Delete: resourceLDAPGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceLDAPGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"dn": {
				Type:        schema.TypeString,
				Description: "The Distinguished Name (DN) of the LDAP group.",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description for the LDAP group.",
				Optional:    true,
			},
			"gid_number": {
				Type:        schema.TypeInt,
				Description: "The numeric group ID for the posixGroup object class.",
				Optional:    true,
			},
			"member": {
				Type:        schema.TypeSet,
				Description: "A list of distinguished names (DNs) that are members of the groupOfNames.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"member_uid": {
				Type:        schema.TypeSet,
				Description: "A list of user IDs (UIDs) that are members of the posixGroup.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"unique_member": {
				Type:        schema.TypeSet,
				Description: "A list of distinguished names (DNs) that are members of the groupOfUniqueNames.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"member_url": {
				Type:        schema.TypeSet,
				Description: "A list of LDAP URLs that can dynamically generate the member list for the groupOfURLs.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"attributes": {
				Type:        schema.TypeSet,
				Description: "The map of attributes of this object; each attribute can be multi-valued.",
				Set:         attributeHash,
				MinItems:    0,

				Elem: &schema.Schema{
					Type:        schema.TypeMap,
					Description: "The list of values for a given attribute.",
					MinItems:    1,
					MaxItems:    1,
					Elem: &schema.Schema{
						Type:        schema.TypeString,
						Description: "The individual value for the given attribute.",
					},
				},
				Optional: true,
			},
			"object_classes": {
				Type:        schema.TypeSet,
				Description: "List of object class names to be used for the LDAP group",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceLDAPGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_group::create - creating a new group with DN %q", dn)

	request := ldap.NewAddRequest(dn, []ldap.Control{})

	// Object class
	var objectClasses []string
	if v, ok := d.GetOk("object_classes"); ok {
		for _, oc := range (v.(*schema.Set)).List() {
			log.Printf("[DEBUG] ldap_object::create - object %q has class: %q", dn, oc.(string))
			objectClasses = append(objectClasses, oc.(string))
		}
	} else {
		// Apply the default value for object_classes
		objectClasses = []string{"posixGroup"}
	}

	errors := validateObjectClasses(objectClasses, d)
	if len(errors) > 0 {
		return fmt.Errorf("validation failed for DN %q: %v", dn, errors)
	}

	request.Attribute("objectClass", objectClasses)

	// Derive the CN (common name) from the DN
	cn, err := deriveCNFromDN(dn)
	if err != nil {
		// Handle Error
		log.Printf("[ERROR] failed to derive CN from DN %q: %v", dn, err)
		return err
	}
	request.Attribute("cn", []string{cn})

	// Description is optional; only set it if provided.
	if description, ok := d.GetOk("description"); ok {
		request.Attribute("description", []string{description.(string)})
	}

	if v, ok := d.GetOk("gid_number"); ok {
		request.Attribute("gidNumber", []string{strconv.Itoa(v.(int))})
	}
	if v, ok := d.GetOk("member_uid"); ok && v.(*schema.Set).Len() > 0 {
		request.Attribute("memberUid", convertToStringSlice(v.(*schema.Set).List()))
	}
	if v, ok := d.GetOk("unique_member"); ok && v.(*schema.Set).Len() > 0 {
		request.Attribute("uniqueMember", convertToStringSlice(v.(*schema.Set).List()))
	}
	if v, ok := d.GetOk("member_url"); ok && v.(*schema.Set).Len() > 0 {
		request.Attribute("memberURL", convertToStringSlice(v.(*schema.Set).List()))
	}
	if v, ok := d.GetOk("member"); ok && v.(*schema.Set).Len() > 0 {
		request.Attribute("member", convertToStringSlice(v.(*schema.Set).List()))
	}
	// if there is a non empty list of attributes, loop though it and
	// create a new map collecting attribute names and its value(s); we need to
	// do this because we could not model the attributes as a map[string][]string
	// due to an appareent limitation in HCL; we have a []map[string]string, so
	// we loop through the list and accumulate values when they share the same
	// key, then we use these as attributes in the LDAP client.
	if v, ok := d.GetOk("attributes"); ok {
		attributes := v.(*schema.Set).List()
		if len(attributes) > 0 {
			log.Printf("[DEBUG] ldap_object::create - object %q has %d attributes", dn, len(attributes))
			m := make(map[string][]string)
			for _, attribute := range attributes {
				log.Printf("[DEBUG] ldap_object::create - %q has attribute of type %T", dn, attribute)
				// each map should only have one entry (see resource declaration)
				for name, value := range attribute.(map[string]interface{}) {
					log.Printf("[DEBUG] ldap_object::create - %q has attribute[%v] => %v (%T)", dn, name, value, value)
					m[name] = append(m[name], value.(string))
				}
			}
			// now loop through the map and add attributes with theys value(s)
			for name, values := range m {
				request.Attribute(name, values)
			}
		}
	}

	// Log the LDAP request attributes before sending the request
	for _, attribute := range request.Attributes {
		log.Printf("[DEBUG] Attribute being added to LDAP request: %s: %+v", attribute.Type, attribute.Vals)
	}
	if err := client.Add(request); err != nil {
		log.Printf("[ERROR] Error while creating LDAP group with DN: %s. Error: %s", dn, err)
		return err
	}

	log.Printf("[DEBUG] ldap_group::create - group %q added to LDAP server", dn)

	d.SetId(dn)                           // The DN is a unique identifier for the group.
	return resourceLDAPGroupRead(d, meta) // Read the new group to update the state.
}

func resourceLDAPGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_group::read - looking for group %q", dn)

	attributesToRead := []string{"cn", "description", "gidNumber", "memberUid", "uniqueMember", "memberURL", "*"}

	request := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectclass=*)",
		attributesToRead,
		nil,
	)

	sr, err := client.Search(request)
	if err != nil {
		if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			log.Printf("[WARN] ldap_group::read - group not found, removing %q from state because it no longer exists in LDAP", dn)
			d.SetId("")
			return nil
		}
		log.Printf("[ERROR] ldap_group::read - lookup for %q failed: %v", dn, err)
		return err
	}

	if len(sr.Entries) == 0 {
		log.Printf("[WARN] ldap_group::read - no results returned for group %q", dn)
		d.SetId("")
		return nil
	}

	entry := sr.Entries[0]
	d.Set("description", entry.GetAttributeValue("description"))
	// Handling gidNumber attribute
	gidNumberStr := entry.GetAttributeValue("gidNumber")
	if gidNumberStr != "" {
		gidNumber, err := strconv.Atoi(gidNumberStr)
		if err != nil {
			log.Printf("[ERROR] unable to convert gidNumber to int: %s", err)
			return err
		}
		d.Set("gid_number", gidNumber)
	}

	// Reading and setting objectClass attribute
	objectClasses := entry.GetAttributeValues("objectClass")
	if len(objectClasses) > 0 {
		d.Set("object_classes", schema.NewSet(schema.HashString, convertToInterfaceSlice(objectClasses)))
	}

	// Reading and setting memberUid attribute
	memberUids := entry.GetAttributeValues("memberUid")
	if len(memberUids) > 0 {
		d.Set("member_uid", memberUids)
	}

	// Reading and setting uniqueMember attribute
	uniqueMembers := entry.GetAttributeValues("uniqueMember")
	if len(uniqueMembers) > 0 {
		d.Set("unique_member", uniqueMembers)
	}

	// Reading and setting memberURL attribute
	memberURLs := entry.GetAttributeValues("memberURL")
	if len(memberURLs) > 0 {
		d.Set("member_url", memberURLs)
	}

	// Reading and setting member attribute
	members := entry.GetAttributeValues("member")
	if len(members) > 0 {
		d.Set("member", memberURLs)
	}
	// Handle other custom attributes
	set := &schema.Set{
		F: attributeHash,
	}
	for _, attribute := range sr.Entries[0].Attributes {
		log.Printf("[DEBUG] ldap_object::read - treating attribute %q of %q (%d values: %v)", attribute.Name, dn, len(attribute.Values), attribute.Values)

		// Skip already-handled or system attributes
		if attribute.Name == "objectClass" || attribute.Name == "cn" || attribute.Name == "description" ||
			attribute.Name == "gidNumber" || attribute.Name == "memberUid" || attribute.Name == "uniqueMember" ||
			attribute.Name == "memberURL" || attribute.Name == "member" {
			log.Printf("[DEBUG] ldap_object::read - skipping attribute %q of %q", attribute.Name, dn)
			continue
		}
		if len(attribute.Values) == 1 {
			// we don't treat the RDN as an ordinary attribute
			a := fmt.Sprintf("%s=%s", attribute.Name, attribute.Values[0])
			if strings.HasPrefix(dn, a) {
				log.Printf("[DEBUG] ldap_object::read - skipping RDN %q of %q", a, dn)
				continue
			}
		}
		log.Printf("[DEBUG] ldap_object::read - adding attribute %q to %q (%d values)", attribute.Name, dn, len(attribute.Values))
		// now add each value as an individual entry into the object, because
		// we do not handle name => []values, and we have a set of maps each
		// holding a single entry name => value; multiple maps may share the
		// same key.
		for _, value := range attribute.Values {
			log.Printf("[DEBUG] ldap_object::read - for %q, setting %q => %q", dn, attribute.Name, value)
			set.Add(map[string]interface{}{
				attribute.Name: value,
			})
		}
	}

	if err := d.Set("attributes", set); err != nil {
		log.Printf("[WARN] ldap_object::read - error setting LDAP attributes for %q : %v", dn, err)
		return err
	}

	log.Printf("[DEBUG] ldap_group::read - finished reading group %q", dn)
	return nil
}

func resourceLDAPGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_group::update - updating group %q", dn)

	request := ldap.NewModifyRequest(dn, []ldap.Control{})

	// Update description if it has changed.
	if d.HasChange("description") {
		request.Replace("description", []string{d.Get("description").(string)})
	}

	if d.HasChange("gid_number") {
		request.Replace("gidNumber", []string{strconv.Itoa(d.Get("gid_number").(int))})
	}

	// Handle updates for member-like attributes
	if err := updateLDAPAttributeSet(request, d, "member", "member"); err != nil {
		return err
	}
	if err := updateLDAPAttributeSet(request, d, "member_uid", "memberUid"); err != nil {
		return err
	}
	if err := updateLDAPAttributeSet(request, d, "unique_member", "uniqueMember"); err != nil {
		return err
	}
	if err := updateLDAPAttributeSet(request, d, "member_url", "memberURL"); err != nil {
		return err
	}

	if d.HasChange("attributes") {
		o, n := d.GetChange("attributes")
		log.Printf("[DEBUG] ldap_object::update - \n%s", printAttributes("old attributes map", o))
		log.Printf("[DEBUG] ldap_object::update - \n%s", printAttributes("new attributes map", n))

		added, changed, removed := computeDeltas(o.(*schema.Set), n.(*schema.Set))
		if len(added) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes added", len(added))
			for _, attr := range added {
				request.Changes = append(request.Changes, ldap.Change{
					Operation:    ldap.AddAttribute,
					Modification: attr,
				})
			}
		}
		if len(changed) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes changed", len(changed))
			for _, attr := range changed {
				request.Changes = append(request.Changes, ldap.Change{
					Operation:    ldap.ReplaceAttribute,
					Modification: attr,
				})
			}
		}
		if len(removed) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes removed", len(removed))
			for _, attr := range removed {
				request.Changes = append(request.Changes, ldap.Change{
					Operation:    ldap.DeleteAttribute,
					Modification: attr,
				})
			}
		}
	}

	// Log the LDAP request modifications before sending the request
	for _, change := range request.Changes {
		log.Printf("[DEBUG] Attribute being changed to LDAP request: %s: %+v",
			change.Modification.Type, change.Modification.Vals)
	}

	// Log the LDAP request modifications before sending the request
	for _, change := range request.Changes {
		operation := "" // will hold the LDAP operation as a string
		switch change.Operation {
		case ldap.AddAttribute:
			operation = "Add"
		case ldap.DeleteAttribute:
			operation = "Delete"
		case ldap.ReplaceAttribute:
			operation = "Replace"
		}
		log.Printf("[DEBUG] ModifyRequest Change - Operation: %s, Type: %s, Values: %v",
			operation, change.Modification.Type, change.Modification.Vals)
	}

	if err := client.Modify(request); err != nil {
		log.Printf("[ERROR] ldap_group::update - error updating group %q: %v", dn, err)
		return err
	}

	return resourceLDAPGroupRead(d, meta)
}

func resourceLDAPGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_group::delete - removing group %q", dn)

	request := ldap.NewDelRequest(dn, []ldap.Control{})

	if err := client.Del(request); err != nil {
		log.Printf("[ERROR] ldap_group::delete - error removing group %q: %v", dn, err)
		return err
	}

	log.Printf("[DEBUG] ldap_group::delete - group %q removed", dn)
	d.SetId("") // This will remove the resource from the state file.
	return nil
}

func resourceLDAPGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The DN is the ID
	dn := d.Id()
	d.Set("dn", dn)

	// Call the read function to ensure the data is fully populated
	if err := resourceLDAPGroupRead(d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// deriveCNFromDN is a helper function to extract the CN from a DN.
// It assumes the DN is in the form "CN=groupname,OU=subunit,DC=example,DC=com".
func deriveCNFromDN(dn string) (string, error) {
	parts := strings.SplitN(dn, ",", 2)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid DN format")
	}

	cnPart := parts[0]
	if strings.HasPrefix(strings.ToUpper(cnPart), "CN=") {
		return cnPart[3:], nil // strip the "CN=" prefix and return the CN value
	}

	return "", fmt.Errorf("CN not found in DN")
}

func convertToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

func convertToInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, v := range strings {
		result[i] = v
	}
	return result
}

// validateObjectClasses checks if all required attributes for given objectClasses are available.
func validateObjectClasses(objectClasses []string, d *schema.ResourceData) []error {
	var errors []error
	for _, objectClass := range objectClasses {
		switch objectClass {
		case "posixGroup":
			if _, ok := d.GetOk("gid_number"); !ok {
				errors = append(errors, fmt.Errorf("missing required attribute 'gid_number' for objectClass 'posixGroup'"))
			}
			// ... check other attributes for posixGroup ...
		}

	}
	return errors
}

// updateLDAPAttributeSet handles the update logic for member-like attributes in a DRY manner.
func updateLDAPAttributeSet(request *ldap.ModifyRequest, d *schema.ResourceData, tfAttributeName string, ldapAttributeName string) error {
	if d.HasChange(tfAttributeName) {
		oldVal, newVal := d.GetChange(tfAttributeName)
		oldSet := oldVal.(*schema.Set)
		newSet := newVal.(*schema.Set)

		// obtaining strings to add
		for _, add := range newSet.Difference(oldSet).List() {
			request.Add(ldapAttributeName, []string{add.(string)})
		}

		// obtaining strings to remove
		for _, remove := range oldSet.Difference(newSet).List() {
			request.Delete(ldapAttributeName, []string{remove.(string)})
		}
	}

	return nil
}
