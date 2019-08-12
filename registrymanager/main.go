package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/umg/docker-registry-manager/pkg/catalog"
	"github.com/umg/docker-registry-manager/pkg/db"
	"github.com/umg/docker-registry-manager/pkg/proxy"
	"github.com/umg/docker-registry-manager/pkg/users"
)

func init() {
	_, err := db.Connect()
	if err != nil {
		log.Fatalf("Database Error: %s\n", err.Error())
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target,
		http.StatusTemporaryRedirect)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := fmt.Sprintf("Contact %s for access.", os.Getenv("ADMIN_EMAIL"))
		fmt.Fprint(w, index)
		return
	})
	r.HandleFunc("/user", users.GetHandler).Methods("GET")
	r.HandleFunc("/user", users.CreateHandler).Methods("POST")
	r.HandleFunc("/user", users.UpdateHandler).Methods("PATCH")
	r.HandleFunc("/user", users.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/users", users.ListHandler).Methods("GET")
	r.HandleFunc("/user/password", users.PasswordChangeHandler).Methods("POST")
	r.HandleFunc("/user/namespaces", users.ChangeNamespacesHandler).Methods("POST")
	r.HandleFunc("/v2/_catalog", catalog.Handler).Methods("GET")
	r.PathPrefix("/").HandlerFunc(proxy.Registry)
	if os.Getenv("CERT_PATH") != "" && os.Getenv("CERT_KEY_PATH") != "" {
		fmt.Println("Listening on port :443")
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))
		err := http.ListenAndServeTLS(":443", os.Getenv("CERT_PATH"), os.Getenv("CERT_KEY_PATH"), r)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Listening on port :80")
		err := http.ListenAndServe(":80", r)
		if err != nil {
			log.Fatal(err)
		}
	}
}
