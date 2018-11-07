package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/umg/docker-registry-manager/pkg/catalog"
	"github.com/umg/docker-registry-manager/pkg/pass"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Must authenticate with BASIC AUTH")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", defaultHandler).Methods("GET")
	r.HandleFunc("/password", pass.PasswordChangeHandler).Methods("POST")
	r.HandleFunc("/catalog", catalog.CatalogHandler).Methods("GET")
	http.ListenAndServe(":8081", r)
}
