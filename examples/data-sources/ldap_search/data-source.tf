# Search for all users in the directory
data "ldap_search" "all_users" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=inetOrgPerson)"
  scope   = "sub"
}

# Output the DNs of all found entries
output "user_dns" {
  value = [for entry in data.ldap_search.all_users.entries : entry.dn]
}

# Search for a specific user by UID
data "ldap_search" "specific_user" {
  base_dn = "dc=example,dc=com"
  filter  = "(uid=johndoe)"
  scope   = "sub"
}

# Search only direct children (one level)
data "ldap_search" "direct_children" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=*)"
  scope   = "one"
}

# Search with specific attributes only
data "ldap_search" "users_email_only" {
  base_dn    = "ou=users,dc=example,dc=com"
  filter     = "(objectClass=inetOrgPerson)"
  scope      = "sub"
  attributes = ["mail", "cn", "uid"]
}

# Search for a single object (base scope)
data "ldap_search" "single_object" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=organizationalUnit)"
  scope   = "base"
}

# Example: Extract email addresses from search results
output "user_emails" {
  value = [
    for entry in data.ldap_search.users_email_only.entries :
    lookup(entry.attributes, "mail", "")
  ]
}
