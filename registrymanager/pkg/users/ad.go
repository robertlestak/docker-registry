package users

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/ldap.v3"
)

// Connect connects to the LDAP server and returns the connection
func Connect() (*ldap.Conn, error) {
	p, perr := strconv.Atoi(os.Getenv("LDAP_PORT"))
	if perr != nil {
		return nil, perr
	}
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", os.Getenv("LDAP_SERVER"), p))
	if err != nil {
		return l, err
	}
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return l, err
	}
	return l, nil
}

// Authenticate authenticates a user in LDAP
func Authenticate(l *ldap.Conn, u *User) (bool, error) {
	err := l.Bind(u.LDAPDN, u.Password)
	if err != nil {
		return false, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	err = l.Bind(os.Getenv("LDAP_SVC_USER"), os.Getenv("LDAP_SVC_PASS"))
	if err != nil {
		return false, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return true, nil
}

// Search searches LDAP for a user and returns DN if the user exists
func Search(l *ldap.Conn, u *User) (string, error) {
	var dn string
	err := l.Bind(os.Getenv("LDAP_SVC_USER"), os.Getenv("LDAP_SVC_PASS"))
	if err != nil {
		return dn, err
	}
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(%s=%s))", os.Getenv("LDAP_SEARCH_KEY"), u.Username),
		[]string{"dn"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return dn, err
	}
	if len(sr.Entries) != 1 {
		return dn, errors.New("User does not exist or too many entries returned")
	}
	u.LDAPDN = sr.Entries[0].DN
	return u.LDAPDN, nil
}

// ADAuth returns true if the user exists in AD
func (u *User) ADAuth() (bool, error) {
	if os.Getenv("LDAP_SERVER") == "" {
		return false, errors.New("LDAP server required")
	}
	l, err := Connect()
	if err != nil {
		return false, err
	}
	defer l.Close()
	_, lerr := Search(l, u)
	if lerr != nil {
		return false, lerr
	}
	auth, aerr := Authenticate(l, u)
	if aerr != nil {
		return false, aerr
	}
	return auth, nil
}
