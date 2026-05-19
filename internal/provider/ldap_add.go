package provider

import (
	"log"

	"github.com/go-ldap/ldap/v3"
)

// addLDAPEntry adds an LDAP entry and retries referrals with the ManageDsaIT
// control. Some directory servers return LDAPResultReferral when creating an
// entry below a referral-managed DSA boundary. Retrying with ManageDsaIT lets
// the server treat referral objects as normal entries, matching the delete
// referral handling while keeping the normal add path unchanged.
func addLDAPEntry(conn *ldap.Conn, request *ldap.AddRequest, dn string, logPrefix string) error {
	if err := conn.Add(request); err != nil {
		ldapErr, ok := err.(*ldap.Error)
		if !ok {
			log.Printf("[ERROR] %s - error adding %q: %v", logPrefix, dn, err)
			return err
		}

		switch ldapErr.ResultCode {
		case ldap.LDAPResultReferral:
			log.Printf("[WARN] %s - add of %q returned referral, retrying with ManageDsaIT control", logPrefix, dn)
			return addLDAPEntryWithManageDsaIT(conn, request, dn, logPrefix)
		default:
			log.Printf("[ERROR] %s - error adding %q: %v", logPrefix, dn, err)
			return err
		}
	}

	return nil
}

func addLDAPEntryWithManageDsaIT(conn *ldap.Conn, request *ldap.AddRequest, dn string, logPrefix string) error {
	request.Controls = append(request.Controls, ldap.NewControlManageDsaIT(false))

	if err := conn.Add(request); err != nil {
		log.Printf("[ERROR] %s - error adding %q with ManageDsaIT control: %v", logPrefix, dn, err)
		return err
	}

	return nil
}
