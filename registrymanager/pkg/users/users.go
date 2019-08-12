package users

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/umg/docker-registry-manager/pkg/db"
)

// User handles user data
type User struct {
	ID          int       `json:"id,omitempty"`
	Username    string    `json:"username"`
	Password    string    `json:"password,omitempty"`
	AD          bool      `json:"ad"`
	LDAPDN      string    `json:"ldap_dn,omitempty"`
	Admin       bool      `json:"admin"`
	NewPassword string    `json:"-"`
	Namespaces  []string  `json:"namespaces,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func trimStringSlice(s []string) []string {
	var ns []string
	for _, v := range s {
		nv := strings.TrimSpace(v)
		if nv != "" {
			ns = append(ns, nv)
		}
	}
	return ns
}

// Authenticated checks the database to determine if a user is authorized
func (u *User) Authenticated() (bool, error) {
	var auth bool
	var e error
	if u.Username == "" || u.Password == "" {
		return auth, errors.New("Username and Password required")
	}
	isAd, iaerr := u.isAD()
	if iaerr != nil {
		return false, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	if isAd {
		ad, aerr := u.ADAuth()
		if aerr != nil {
			return false, errors.New(http.StatusText(http.StatusUnauthorized))
		}
		return ad, nil
	}
	fmt.Printf("%+v\n", u)
	qry := `SELECT id FROM users
            WHERE username=$1 AND
            password=crypt($2, password)
						AND deleted_at IS NULL
          `
	err := db.DB.QueryRow(qry, u.Username, u.Password).Scan(&u.ID)
	if err != nil {
		return auth, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	if u.ID > 0 {
		auth = true
	}
	return auth, e
}

func (u *User) isAD() (bool, error) {
	if u.Username == "" {
		return false, errors.New("Username required")
	}
	qry := `SELECT ad FROM users
						WHERE username=$1
						AND deleted_at IS NULL
          `
	err := db.DB.QueryRow(qry, u.Username).Scan(&u.AD)
	if err != nil {
		return u.AD, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return u.AD, nil
}

// Get returns the accounts and tags for a user
func (u *User) Get() error {
	var e error
	if u.Username == "" {
		return errors.New("Username required")
	}
	qry := `SELECT ad, admin, namespaces, created_at FROM users
            WHERE username=$1
						AND deleted_at IS NULL
          `
	err := db.DB.QueryRow(qry, u.Username).Scan(&u.AD, &u.Admin, pq.Array(&u.Namespaces), &u.CreatedAt)
	if err != nil {
		return errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return e
}

// List returns the accounts and tags for users
func List(o int, l int) ([]User, error) {
	var e error
	qry := `SELECT username, ad, admin, namespaces, created_at FROM users
						WHERE deleted_at IS NULL
						OFFSET $1 LIMIT $2
          `
	var ul []User
	if l > 50 {
		l = 50
	}
	rows, err := db.DB.Query(qry, o, l)
	if err != nil {
		return ul, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	for rows.Next() {
		var u User
		serr := rows.Scan(&u.Username, &u.AD, &u.Admin, pq.Array(&u.Namespaces), &u.CreatedAt)
		if serr != nil {
			return ul, errors.New(http.StatusText(http.StatusUnauthorized))
		}
		ul = append(ul, u)
	}
	return ul, e
}

// GetNamespaces returns the accounts and tags for a user
func (u *User) GetNamespaces() error {
	var e error
	auth, aerr := u.Authenticated()
	if aerr != nil {
		return aerr
	} else if !auth {
		return errors.New(http.StatusText(http.StatusUnauthorized))
	}
	qry := `SELECT namespaces FROM users
            WHERE username=$1
						AND deleted_at IS NULL
          `
	err := db.DB.QueryRow(qry, u.Username).Scan(pq.Array(&u.Namespaces))
	if err != nil {
		return errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return e
}

// Create inserts the specified User into the system
func (u *User) Create() (*User, error) {
	var e error
	if u.Username == "" {
		return u, errors.New(http.StatusText(http.StatusBadRequest))
	}
	if u.AD {
		ex, aderr := u.ADUserExists()
		if aderr != nil {
			return u, aderr
		}
		if !ex {
			return u, errors.New(http.StatusText(http.StatusNetworkAuthenticationRequired))
		}
	}
	u.Namespaces = trimStringSlice(u.Namespaces)
	qry := `INSERT INTO users
            (username, password, ad, admin, namespaces) VALUES
            ($1, crypt($2, gen_salt('bf')), $3, $4, $5)
          RETURNING id, created_at
          `
	err := db.DB.QueryRow(qry, u.Username, u.Password, u.AD, u.Admin, pq.Array(u.Namespaces)).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return u, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return u, e
}

// Delete removes the user specified by username
func (u *User) Delete() error {
	var e error
	qry := `UPDATE users SET deleted_at=current_timestamp
						WHERE username=$1
						AND deleted_at IS NULL
					RETURNING id
				 `
	err := db.DB.QueryRow(qry, u.Username).Scan(&u.ID)
	if err != nil {
		return errors.New(http.StatusText(http.StatusUnauthorized))
	}
	return e
}

// ChangePass checks the password of the existing user
// and changes the password to NewPassword if the current password is correct
func (u *User) ChangePass() (*User, error) {
	var e error
	qry := `UPDATE users SET
						password=crypt($2, gen_salt('bf'))
						WHERE username=$1
					RETURNING id, created_at
				 `
	err := db.DB.QueryRow(qry, u.Username, u.NewPassword).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return u, errors.New(http.StatusText(http.StatusBadRequest))
	}
	return u, e
}

// UpdateNamespaces changes the accounts for a user.
func (u *User) UpdateNamespaces() (*User, error) {
	var e error
	qry := `UPDATE users SET namespaces = $1 WHERE username = $2
	          `
	u.Namespaces = trimStringSlice(u.Namespaces)
	_, err := db.DB.Exec(qry, pq.Array(u.Namespaces), u.Username)
	if err != nil {
		return u, errors.New(http.StatusText(http.StatusBadRequest))
	}
	return u, e
}

// UpdateUser allows the admin user to update a user object
func (u *User) UpdateUser() (*User, error) {
	var e error
	qry := `UPDATE users SET
						password=crypt($2, gen_salt('bf')),
						ad=$3, admin=$4, namespaces=$5
						WHERE username=$1
					RETURNING id, created_at
				 `
	u.Namespaces = trimStringSlice(u.Namespaces)
	err := db.DB.QueryRow(qry, u.Username, u.Password, u.AD, u.Admin, pq.Array(u.Namespaces)).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return u, errors.New(http.StatusText(http.StatusBadRequest))
	}
	return u, e
}
