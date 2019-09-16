package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/umg/docker-registry-manager/pkg/catalog"
	"github.com/umg/docker-registry-manager/pkg/db"
	"github.com/umg/docker-registry-manager/pkg/proxy"
	"github.com/umg/docker-registry-manager/pkg/users"
)

// BasicAuthRealm is the string name of the realm
const BasicAuthRealm string = "Docker Registry"

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

func getCertsInDir(d string) map[string]string {
	crts, err := ioutil.ReadDir(d)
	if err != nil {
		log.Fatal(err)
	}
	crtPair := make(map[string]string)
	for _, c := range crts {
		if strings.HasSuffix(c.Name(), ".key") {
			crtPair["key"] = path.Join(d, c.Name())
		} else if strings.HasSuffix(c.Name(), ".pem") {
			crtPair["pem"] = path.Join(d, c.Name())
		}
	}
	return crtPair
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

	if os.Getenv("CERTS_DIR") != "" {
		var crtPairs []map[string]string
		crtDir, cerr := ioutil.ReadDir(os.Getenv("CERTS_DIR"))
		if cerr != nil {
			log.Fatal(cerr)
		}
		for _, f := range crtDir {
			if f.IsDir() {
				crtPair := getCertsInDir(path.Join(os.Getenv("CERTS_DIR"), f.Name()))
				crtPairs = append(crtPairs, crtPair)
			} else {
				crtPair := getCertsInDir(os.Getenv("CERTS_DIR"))
				crtPairs = append(crtPairs, crtPair)
			}
		}
		var err error
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = make([]tls.Certificate, len(crtPairs))
		for i, c := range crtPairs {
			tlsConfig.Certificates[i], err = tls.LoadX509KeyPair(c["pem"], c["key"])
			if err != nil {
				log.Fatal(err)
			}
		}
		tlsConfig.BuildNameToCertificate()
		fmt.Println("Listening on port :443")
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))
		server := &http.Server{
			Handler:   r,
			TLSConfig: tlsConfig,
		}
		listener, err := tls.Listen("tcp", ":443", tlsConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(server.Serve(listener))
	} else {
		fmt.Println("Listening on port :80")
		err := http.ListenAndServe(":80", r)
		if err != nil {
			log.Fatal(err)
		}
	}
}
