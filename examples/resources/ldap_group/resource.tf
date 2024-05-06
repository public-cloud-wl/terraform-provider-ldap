resource "ldap_object" "users_example_com" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["top", "organizationalUnit"]
}

resource "ldap_object" "a123456" {
  dn             = "uid=a123456,${ldap_object.users_example_com.dn}"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "Doe" },
    { givenName = "John" },
    { cn = "John Doe" },
    { displayName = "Mr. John K. Doe, esq." },
    { mail = "john.doe@example.com" },
    { mail = "jdoe@example.com" },
    { userPassword = "password" },
    { uidNumber = "1234" },
    { gidNumber = "1234" },
    { homeDirectory = "/home/jdoe" },
    { loginShell = "/bin/bash" }
  ]
}

resource "ldap_object" "groups_example_com" {
  dn             = "ou=groups,dc=example,dc=com"
  object_classes = ["top", "organizationalUnit"]
}

resource "ldap_group" "my_group" {
  dn          = "cn=mygroup,${ldap_object.users_example_com.dn}"
  object_classes = ["top", "posixGroup", "groupOfUniqueNames", "groupOfURLs"]
  description = "My Group"
  gid_number   = "1234567890"
  attributes  = [{
    businessCategory = "My Business Category"
  }]
  unique_member = [ ldap_object.a123456.id ]
}