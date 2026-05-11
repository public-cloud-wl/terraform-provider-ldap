# Keyed map of users — O(1) lookup by uid
data "ldap_search_map" "users" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(objectClass=inetOrgPerson)"
  scope         = "sub"
  key_attribute = "uid"
}

output "alice_dn" {
  value = data.ldap_search_map.users.dns["alice"]
}

output "alice_attributes" {
  value = jsondecode(data.ldap_search_map.users.attributes_json_by_key["alice"])
}

# Large directory — enable pagination and limit fetched attributes
data "ldap_search_map" "groups" {
  base_dn              = "ou=groups,dc=example,dc=com"
  filter               = "(objectClass=groupOfNames)"
  scope                = "sub"
  key_attribute        = "cn"
  paged_size           = 500
  requested_attributes = ["cn", "member"]
}
