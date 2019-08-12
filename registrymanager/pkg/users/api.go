package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// GetCurrent returns the user for the current request
func GetCurrent(r *http.Request) (*User, error) {
	un, pass, _ := r.BasicAuth()
	u := &User{
		Username: un,
		Password: pass,
	}
	if u.Username == os.Getenv("REGISTRY_ADMIN_USER") && u.Password == os.Getenv("REGISTRY_ADMIN_PASS") {
		u.Admin = true
		return u, nil
	}
	auth, aerr := u.Authenticated()
	if aerr != nil {
		return u, aerr
	}
	if !auth {
		return u, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	err := u.Get()
	if err != nil {
		return u, err
	}
	return u, nil
}

// reqIsAdmin gets the current request user and returns true if they are an admin
func reqIsAdmin(r *http.Request) bool {
	u, err := GetCurrent(r)
	if err != nil {
		return false
	}
	return u.Admin
}

// CreateHandler creates a user
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if !reqIsAdmin(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	var bad bool
	var badmin bool
	if r.FormValue("ad") != "" {
		var err error
		bad, err = strconv.ParseBool(r.FormValue("ad"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.FormValue("admin") != "" {
		var err error
		badmin, err = strconv.ParseBool(r.FormValue("admin"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	var pass string
	if !bad {
		pass = r.FormValue("password")
	}
	u := &User{
		Username:   r.FormValue("username"),
		Password:   pass,
		AD:         bad,
		Admin:      badmin,
		Namespaces: strings.Split(r.FormValue("namespaces"), ","),
	}
	var err error
	u, err = u.Create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, "User created")
}

// PasswordChangeHandler changes a password for a user
func PasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	u := &User{
		Username:    un,
		Password:    p,
		NewPassword: r.FormValue("password"),
	}
	auth, aerr := u.Authenticated()
	if aerr != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if !auth {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if reqIsAdmin(r) && r.FormValue("username") != "" {
		u.Username = r.FormValue("username")
	} else if !reqIsAdmin(r) && r.FormValue("username") != "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	uerr := u.Get()
	if uerr != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if u.AD {
		http.Error(w, http.StatusText(http.StatusNetworkAuthenticationRequired), http.StatusBadRequest)
		return
	}
	u, err := u.ChangePass()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "Password Changed\n")
}

// UpdateHandler enables an admin to change a user's password
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	if !reqIsAdmin(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	var bad bool
	var badmin bool
	if r.FormValue("ad") != "" {
		var err error
		bad, err = strconv.ParseBool(r.FormValue("ad"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.FormValue("admin") != "" {
		var err error
		badmin, err = strconv.ParseBool(r.FormValue("admin"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	u := &User{
		Username:   r.FormValue("username"),
		Password:   r.FormValue("password"),
		Admin:      badmin,
		AD:         bad,
		Namespaces: strings.Split(r.FormValue("namespaces"), ","),
	}
	u, uerr := u.UpdateUser()
	if uerr != nil {
		http.Error(w, uerr.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "User updated\n")
}

// DeleteHandler enables an admin to delete a user
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	if !reqIsAdmin(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	u := &User{
		Username: r.FormValue("username"),
	}
	err := u.Delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "User Deleted\n")
}

// ChangeNamespacesHandler enables an admin to change a user's namespace
func ChangeNamespacesHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	if !reqIsAdmin(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	u := &User{
		Username:   r.FormValue("username"),
		Namespaces: strings.Split(r.FormValue("namespaces"), ","),
	}
	u, err := u.UpdateNamespaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "Namespaces Changed\n")
}

// GetHandler returns the data for a user
func GetHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	u := &User{
		Username: un,
	}
	if reqIsAdmin(r) && r.FormValue("username") != "" {
		u.Username = r.FormValue("username")
	}
	err := u.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jd, jerr := json.Marshal(&u)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, string(jd))
}

// ListHandler enables an admin to list all users
func ListHandler(w http.ResponseWriter, r *http.Request) {
	un, p, _ := r.BasicAuth()
	if un == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	if !reqIsAdmin(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	off := 0
	lim := 50
	var of string
	var li string
	var err error
	if r.FormValue("offset") != "" {
		of = r.FormValue("offset")
		off, err = strconv.Atoi(of)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.FormValue("limit") != "" {
		li = r.FormValue("limit")
		lim, err = strconv.Atoi(li)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	usrs, err := List(off, lim)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jd, jerr := json.Marshal(&usrs)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, string(jd))
}
