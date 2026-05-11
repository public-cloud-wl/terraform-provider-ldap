package provider

import (
	"log"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// deleteLDAPEntry deletes an LDAP entry and retries referrals with the ManageDsaIT
// control. Some directory servers return LDAPResultReferral when deleting referral
// entries or entries below a referral; LDAP admin clients commonly send ManageDsaIT
// for these operations. Terraform should be able to do the same while keeping the
// normal delete path unchanged for regular entries.
func deleteLDAPEntry(conn *ldap.Conn, dn string, logPrefix string) error {
	request := ldap.NewDelRequest(dn, []ldap.Control{})

	if err := conn.Del(request); err != nil {
		ldapErr, ok := err.(*ldap.Error)
		if !ok {
			log.Printf("[ERROR] %s - error removing %q: %v", logPrefix, dn, err)
			return err
		}

		switch ldapErr.ResultCode {
		case ldap.LDAPResultNoSuchObject:
			log.Printf("[WARN] %s - %q does not exist in LDAP, considering delete successful", logPrefix, dn)
			return nil
		case ldap.LDAPResultReferral:
			log.Printf("[WARN] %s - delete of %q returned referral, retrying with ManageDsaIT control", logPrefix, dn)
			return deleteLDAPEntryWithManageDsaIT(conn, dn, logPrefix)
		case ldap.LDAPResultUnwillingToPerform:
			if isCannotDeleteReferralError(err) {
				log.Printf("[WARN] %s - delete of %q returned unwillingToPerform/cannot delete referral, retrying with ManageDsaIT control", logPrefix, dn)
				return deleteLDAPEntryWithManageDsaIT(conn, dn, logPrefix)
			}
			log.Printf("[ERROR] %s - error removing %q: %v", logPrefix, dn, err)
			return err
		default:
			log.Printf("[ERROR] %s - error removing %q: %v", logPrefix, dn, err)
			return err
		}
	}

	return nil
}

func isCannotDeleteReferralError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "cannot delete referral")
}

func deleteLDAPEntryWithManageDsaIT(conn *ldap.Conn, dn string, logPrefix string) error {
	request := ldap.NewDelRequest(dn, []ldap.Control{ldap.NewControlManageDsaIT(false)})

	if err := conn.Del(request); err != nil {
		if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
			log.Printf("[WARN] %s - %q does not exist in LDAP after ManageDsaIT retry, considering delete successful", logPrefix, dn)
			return nil
		}

		log.Printf("[ERROR] %s - error removing %q with ManageDsaIT control: %v", logPrefix, dn, err)
		return err
	}

	return nil
}
