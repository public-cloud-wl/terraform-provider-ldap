package provider

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceLDAPSearch() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLDAPSearchRead,

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
			"attributes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Specific attributes to retrieve. Default: all attributes.",
			},
			"entries": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of LDAP entries matching the search criteria.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Distinguished Name of the entry.",
						},
						"attributes": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Map of attribute names to their values (comma-separated for multi-valued).",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		Description: "Searches for LDAP objects matching a filter within a base DN.",
	}
}

func dataSourceLDAPSearchRead(d *schema.ResourceData, meta interface{}) error {
	// 1. Extract parameters
	baseDN := d.Get("base_dn").(string)
	filter := d.Get("filter").(string)
	scopeStr := d.Get("scope").(string)

	// 2. Convert scope string to LDAP scope constant
	scope := ldap.ScopeWholeSubtree // default
	switch scopeStr {
	case "base":
		scope = ldap.ScopeBaseObject
	case "one":
		scope = ldap.ScopeSingleLevel
	case "sub":
		scope = ldap.ScopeWholeSubtree
	}

	// 3. Get attributes to retrieve
	attributes := []string{"*"} // default: all
	if v, ok := d.GetOk("attributes"); ok {
		attrList := v.([]interface{})
		attributes = make([]string, len(attrList))
		for i, attr := range attrList {
			attributes[i] = attr.(string)
		}
	}

	// 4. Get LDAP connection
	providerConfig := meta.(*ProviderConfig)
	conn := providerConfig.Connection

	// 5. Build and execute search request
	request := ldap.NewSearchRequest(
		baseDN,
		scope,
		ldap.NeverDerefAliases,
		0,     // no size limit
		0,     // no time limit
		false, // return attribute values, not just names
		filter,
		attributes,
		nil, // no controls
	)

	log.Printf("[DEBUG] ldap_search::read - searching base_dn=%q, filter=%q, scope=%d", baseDN, filter, scope)

	sr, err := conn.Search(request)
	if err != nil {
		return fmt.Errorf("LDAP search failed: %w", err)
	}

	log.Printf("[DEBUG] ldap_search::read - found %d entries", len(sr.Entries))

	// 6. Convert results to Terraform schema format
	entries := make([]interface{}, len(sr.Entries))
	for i, entry := range sr.Entries {
		attrs := make(map[string]interface{})
		for _, attr := range entry.Attributes {
			// Join multi-valued attributes with comma
			attrs[attr.Name] = strings.Join(attr.Values, ",")
		}

		entries[i] = map[string]interface{}{
			"dn":         entry.DN,
			"attributes": attrs,
		}
	}

	// 7. Set computed values
	if err := d.Set("entries", entries); err != nil {
		return fmt.Errorf("error setting entries: %w", err)
	}

	// 8. Generate stable ID for state tracking
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(baseDN+filter+scopeStr)))
	d.SetId(id)

	return nil
}
