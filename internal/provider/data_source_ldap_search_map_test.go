package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLDAPSearchMap_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchMapConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "entry_count", "1"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "dns.testuser", "uid=testuser,ou=users,dc=example,dc=com"),
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "attributes_json_by_key.testuser"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearchMap_multipleEntries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchMapConfig_multiple,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "entry_count", "2"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "dns.testuser1", "uid=testuser1,ou=users,dc=example,dc=com"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "dns.testuser2", "uid=testuser2,ou=users,dc=example,dc=com"),
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "attributes_json_by_key.testuser1"),
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "attributes_json_by_key.testuser2"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearchMap_emptyResults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchMapConfig_emptyResults,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "entry_count", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearchMap_requestedAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchMapConfig_requestedAttributes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "entry_count", "1"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "dns.testuser", "uid=testuser,ou=users,dc=example,dc=com"),
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "attributes_json_by_key.testuser"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearchMap_noPaging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchMapConfig_noPaging,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search_map.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "entry_count", "1"),
					resource.TestCheckResourceAttr("data.ldap_search_map.test", "dns.testuser", "uid=testuser,ou=users,dc=example,dc=com"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearchMap_duplicateKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceLDAPSearchMapConfig_duplicateKey,
				ExpectError: regexp.MustCompile(`duplicate values found for key_attribute`),
			},
		},
	})
}

const testAccDataSourceLDAPSearchMapConfig_basic = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

resource "ldap_object" "test_user" {
  dn             = "uid=testuser,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "User" },
    { cn = "Test User" },
    { uid = "testuser" },
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search_map" "test" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(uid=testuser)"
  scope         = "sub"
  key_attribute = "uid"

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchMapConfig_multiple = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

resource "ldap_object" "test_user1" {
  dn             = "uid=testuser1,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "User1" },
    { cn = "Test User 1" },
    { uid = "testuser1" },
    { uidNumber = "5001" },
    { gidNumber = "5001" },
    { homeDirectory = "/home/testuser1" },
  ]

  depends_on = [ldap_object.users_ou]
}

resource "ldap_object" "test_user2" {
  dn             = "uid=testuser2,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "User2" },
    { cn = "Test User 2" },
    { uid = "testuser2" },
    { uidNumber = "5002" },
    { gidNumber = "5002" },
    { homeDirectory = "/home/testuser2" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search_map" "test" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(objectClass=inetOrgPerson)"
  scope         = "sub"
  key_attribute = "uid"

  depends_on = [ldap_object.test_user1, ldap_object.test_user2]
}
`

const testAccDataSourceLDAPSearchMapConfig_emptyResults = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

data "ldap_search_map" "test" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(uid=nonexistentuser)"
  scope         = "sub"
  key_attribute = "uid"

  depends_on = [ldap_object.users_ou]
}
`

const testAccDataSourceLDAPSearchMapConfig_requestedAttributes = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

resource "ldap_object" "test_user" {
  dn             = "uid=testuser,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "User" },
    { cn = "Test User" },
    { uid = "testuser" },
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search_map" "test" {
  base_dn               = "ou=users,dc=example,dc=com"
  filter                = "(uid=testuser)"
  scope                 = "sub"
  key_attribute         = "uid"
  requested_attributes  = ["cn", "sn", "uid"]

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchMapConfig_noPaging = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

resource "ldap_object" "test_user" {
  dn             = "uid=testuser,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "User" },
    { cn = "Test User" },
    { uid = "testuser" },
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search_map" "test" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(uid=testuser)"
  scope         = "sub"
  key_attribute = "uid"
  paged_size    = 0

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchMapConfig_duplicateKey = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

resource "ldap_object" "test_user1" {
  dn             = "uid=testuser1,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "Shared" },
    { cn = "Shared Name" },
    { uid = "testuser1" },
    { uidNumber = "5001" },
    { gidNumber = "5001" },
    { homeDirectory = "/home/testuser1" },
  ]

  depends_on = [ldap_object.users_ou]
}

resource "ldap_object" "test_user2" {
  dn             = "uid=testuser2,ou=users,dc=example,dc=com"
  object_classes = ["inetOrgPerson", "posixAccount"]
  attributes = [
    { sn = "Shared" },
    { cn = "Shared Name" },
    { uid = "testuser2" },
    { uidNumber = "5002" },
    { gidNumber = "5002" },
    { homeDirectory = "/home/testuser2" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search_map" "test" {
  base_dn       = "ou=users,dc=example,dc=com"
  filter        = "(objectClass=inetOrgPerson)"
  scope         = "sub"
  key_attribute = "sn"

  depends_on = [ldap_object.test_user1, ldap_object.test_user2]
}
`
