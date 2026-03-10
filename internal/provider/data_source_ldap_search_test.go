package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLDAPSearch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "1"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.0.dn", "uid=testuser,ou=users,dc=example,dc=com"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearch_multipleEntries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_multiple,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearch_scopeOne(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_scopeOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearch_scopeBase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_scopeBase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "1"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.0.dn", "ou=users,dc=example,dc=com"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearch_specificAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_specificAttributes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "1"),
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "entries.0.attributes.cn"),
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "entries.0.attributes.sn"),
				),
			},
		},
	})
}

func TestAccDataSourceLDAPSearch_emptyResults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLDAPSearchConfig_emptyResults,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ldap_search.test", "id"),
					resource.TestCheckResourceAttr("data.ldap_search.test", "entries.#", "0"),
				),
			},
		},
	})
}

const testAccDataSourceLDAPSearchConfig_basic = `
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
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search" "test" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(uid=testuser)"
  scope   = "sub"

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchConfig_multiple = `
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
    { uidNumber = "5002" },
    { gidNumber = "5002" },
    { homeDirectory = "/home/testuser2" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search" "test" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=inetOrgPerson)"
  scope   = "sub"

  depends_on = [ldap_object.test_user1, ldap_object.test_user2]
}
`

const testAccDataSourceLDAPSearchConfig_scopeOne = `
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
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search" "test" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=inetOrgPerson)"
  scope   = "one"

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchConfig_scopeBase = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

data "ldap_search" "test" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(objectClass=organizationalUnit)"
  scope   = "base"

  depends_on = [ldap_object.users_ou]
}
`

const testAccDataSourceLDAPSearchConfig_specificAttributes = `
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
    { uidNumber = "5000" },
    { gidNumber = "5000" },
    { homeDirectory = "/home/testuser" },
  ]

  depends_on = [ldap_object.users_ou]
}

data "ldap_search" "test" {
  base_dn    = "ou=users,dc=example,dc=com"
  filter     = "(uid=testuser)"
  scope      = "sub"
  attributes = ["cn", "sn", "uid"]

  depends_on = [ldap_object.test_user]
}
`

const testAccDataSourceLDAPSearchConfig_emptyResults = `
resource "ldap_object" "users_ou" {
  dn             = "ou=users,dc=example,dc=com"
  object_classes = ["organizationalUnit"]
}

data "ldap_search" "test" {
  base_dn = "ou=users,dc=example,dc=com"
  filter  = "(uid=nonexistentuser)"
  scope   = "sub"

  depends_on = [ldap_object.users_ou]
}
`
