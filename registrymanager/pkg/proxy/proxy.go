package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"

	"github.com/umg/docker-registry-manager/pkg/users"
)

// BasicAuthRealm is the string name of the realm
const BasicAuthRealm string = "Docker Registry"

// Registry autheniticates the user and then forwards requests to the registry
func Registry(w http.ResponseWriter, r *http.Request) {
	_, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, BasicAuthRealm))
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized) + "\n"))
		return
	}
	d := func(req *http.Request) {
		req.URL.Scheme = os.Getenv("REGISTRY_SCHEMA")
		req.URL.Host = os.Getenv("REGISTRY_HOST") + ":" + os.Getenv("REGISTRY_PORT")
		req.Host = os.Getenv("REGISTRY_HOST") + ":" + os.Getenv("REGISTRY_PORT")
		req.URL.Path = r.URL.Path
		req.SetBasicAuth(os.Getenv("REGISTRY_ADMIN_USER"), os.Getenv("REGISTRY_ADMIN_PASS"))
	}
	u, uerr := users.GetCurrent(r)
	if uerr != nil {
		http.Error(w, uerr.Error(), http.StatusUnauthorized)
		return
	}
	if r.URL.Path != "/v2/" && !u.Admin {
		var authed bool
		for _, ns := range u.Namespaces {
			rm := regexp.MustCompile("^/" + os.Getenv("REGISTRY_VERSION") + "/" + ns + "/")
			if rm.MatchString(r.URL.Path) {
				authed = true
			}
		}
		if !authed {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}
	e := func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, e.Error(), http.StatusBadGateway)
	}
	mr := func(r *http.Response) error {
		r.Header.Set("Docker-Distribution-Api-Version", "registry/2.0")
		return nil
	}
	p := &httputil.ReverseProxy{
		Director:       d,
		ErrorHandler:   e,
		ModifyResponse: mr,
	}
	p.ServeHTTP(w, r)
}
