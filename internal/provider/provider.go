package provider

import (
	"fmt"
	"strings"

	"github.com/elastic-infra/terraform-provider-ldap/internal/helper/client"
	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ProviderConfig struct {
	Connection             *ldap.Conn
	InvalidAttributeValues []string
}

// Provider creates a new LDAP provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ldap_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_HOST", nil),
				Description: "The LDAP server to connect to.",
			},
			"ldap_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_PORT", 389),
				Description: "The LDAP protocol port (default: 389).",
			},
			"bind_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_BIND_USER", nil),
				Description: "Bind user to be used for authenticating on the LDAP server.",
			},
			"bind_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_BIND_PASSWORD", nil),
				Description: "Password to authenticate the Bind user.",
			},
			"start_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_START_TLS", false),
				Description: "Upgrade TLS to secure the connection (default: false).",
			},
			"tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_TLS", false),
				Description: "Enable TLS encryption for LDAP (LDAPS) (default: false).",
			},
			"tls_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_TLS_INSECURE", false),
				Description: "Don't verify server TLS certificate (default: false).",
			},
			"invalid_attribute_values": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of attribute values that are considered invalid.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ldap_object": resourceLDAPObject(),
			"ldap_group":  resourceLDAPGroup(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := &client.Config{
		LDAPHost:     d.Get("ldap_host").(string),
		LDAPPort:     d.Get("ldap_port").(int),
		BindUser:     d.Get("bind_user").(string),
		BindPassword: d.Get("bind_password").(string),
		StartTLS:     d.Get("start_tls").(bool),
		TLS:          d.Get("tls").(bool),
		TLSInsecure:  d.Get("tls_insecure").(bool),
	}

	connection, err := client.DialAndBind(config)
	if err != nil {
		return nil, err
	}

	invalidValues := []string{}
	if v, ok := d.GetOk("invalid_attribute_values"); ok && v != nil {
		invalidValues = convertToStringSlice(v.([]interface{}))
	}

	return &ProviderConfig{
		Connection:             connection,
		InvalidAttributeValues: invalidValues,
	}, nil
}

func validateAttributes(d *schema.ResourceData, invalidValues []string) error {
	if v, ok := d.GetOk("attributes"); ok {
		attributes := v.(*schema.Set).List()
		for _, attribute := range attributes {
			for name, value := range attribute.(map[string]interface{}) {
				valStr := value.(string)
				for _, invalid := range invalidValues {
					if strings.EqualFold(valStr, invalid) {
						return fmt.Errorf("attribute %q has invalid value '%s'", name, valStr)
					}
				}
			}
		}
	}
	return nil
}

func convertToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}
