package provider

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLDAPSearchMap() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLDAPSearchMapRead,

		Schema: map[string]*schema.Schema{
			"base_dn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The base DN to start the search from.",
			},
			"filter": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP filter string (e.g., \"(objectClass=inetOrgPerson)\").",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "sub",
				Description:  "Search scope: base, one, or sub. Default: sub.",
				ValidateFunc: validation.StringInSlice([]string{"base", "one", "sub"}, false),
			},
			"requested_attributes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Specific attributes to retrieve. Default: all attributes.",
			},
			"key_attribute": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Attribute whose first value is used as the key of the returned map.",
			},
			"paged_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1000,
				Description: "LDAP paged search size. Set to 0 to disable pagination and use a single search request.",
			},
			"entry_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of LDAP entries matching the search criteria.",
			},
			"dns": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map from key_attribute value to distinguished name.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"attributes_json_by_key": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map from key_attribute value to raw attribute JSON for that entry.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},

		Description: "Searches for LDAP objects and returns them as a map keyed by a chosen LDAP attribute.",
	}
}

func dataSourceLDAPSearchMapRead(d *schema.ResourceData, meta interface{}) error {
	baseDN := d.Get("base_dn").(string)
	filter := d.Get("filter").(string)
	scopeStr := d.Get("scope").(string)
	keyAttribute := d.Get("key_attribute").(string)
	pagedSize := d.Get("paged_size").(int)

	scope := ldap.ScopeWholeSubtree
	switch scopeStr {
	case "base":
		scope = ldap.ScopeBaseObject
	case "one":
		scope = ldap.ScopeSingleLevel
	case "sub":
		scope = ldap.ScopeWholeSubtree
	}

	attributes := []string{"*"}
	if v, ok := d.GetOk("requested_attributes"); ok {
		attrList := v.([]interface{})
		attributes = make([]string, len(attrList))
		for i, attr := range attrList {
			attributes[i] = attr.(string)
		}
	}

	foundKeyAttribute := false
	for _, attr := range attributes {
		if attr == keyAttribute {
			foundKeyAttribute = true
			break
		}
	}
	if !foundKeyAttribute && !(len(attributes) == 1 && attributes[0] == "*") {
		attributes = append(attributes, keyAttribute)
	}

	providerConfig := meta.(*ProviderConfig)
	conn := providerConfig.Connection

	request := ldap.NewSearchRequest(
		baseDN,
		scope,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	log.Printf("[DEBUG] ldap_search_map::read - searching base_dn=%q, filter=%q, scope=%d, key_attribute=%q, paged_size=%d", baseDN, filter, scope, keyAttribute, pagedSize)

	var sr *ldap.SearchResult
	var err error
	if pagedSize > 0 {
		sr, err = conn.SearchWithPaging(request, uint32(pagedSize))
	} else {
		sr, err = conn.Search(request)
	}
	if err != nil {
		return fmt.Errorf("LDAP search failed: %w", err)
	}

	log.Printf("[DEBUG] ldap_search_map::read - found %d entries", len(sr.Entries))

	dns := make(map[string]interface{}, len(sr.Entries))
	attributesJSONByKey := make(map[string]interface{}, len(sr.Entries))
	duplicateKeys := make([]string, 0)

	for _, entry := range sr.Entries {
		keyValues := entry.GetAttributeValues(keyAttribute)
		if len(keyValues) == 0 || keyValues[0] == "" {
			continue
		}
		key := keyValues[0]
		if _, exists := dns[key]; exists {
			duplicateKeys = append(duplicateKeys, key)
			continue
		}

		attrs := make(map[string][]string, len(entry.Attributes))
		for _, attr := range entry.Attributes {
			values := make([]string, len(attr.Values))
			copy(values, attr.Values)
			attrs[attr.Name] = values
		}

		attrsJSON, err := json.Marshal(attrs)
		if err != nil {
			return fmt.Errorf("error marshalling attributes for key %q: %w", key, err)
		}

		dns[key] = entry.DN
		attributesJSONByKey[key] = string(attrsJSON)
	}

	if len(duplicateKeys) > 0 {
		sort.Strings(duplicateKeys)
		return fmt.Errorf("duplicate values found for key_attribute %q: %v", keyAttribute, duplicateKeys)
	}

	if err := d.Set("entry_count", len(sr.Entries)); err != nil {
		return fmt.Errorf("error setting entry_count: %w", err)
	}
	if err := d.Set("dns", dns); err != nil {
		return fmt.Errorf("error setting dns: %w", err)
	}
	if err := d.Set("attributes_json_by_key", attributesJSONByKey); err != nil {
		return fmt.Errorf("error setting attributes_json_by_key: %w", err)
	}

	id := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", baseDN, filter, scopeStr, keyAttribute, pagedSize))))
	d.SetId(id)

	return nil
}

