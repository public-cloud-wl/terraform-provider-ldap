# Ordered list of all users — iterate by index or use for_each
data "ldap_search" "users" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=inetOrgPerson)"
  scope   = "sub"
}

output "user_dns" {
  value = [for e in data.ldap_search.users.entries : e.dn]
}

# Limit fetched attributes
data "ldap_search" "users_slim" {
  base_dn    = "ou=users,dc=example,dc=com"
  filter     = "(objectClass=inetOrgPerson)"
  scope      = "sub"
  attributes = ["cn", "mail"]
}
