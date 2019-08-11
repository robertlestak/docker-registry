package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/umg/docker-registry-manager/pkg/users"
)

// Catalog contains the repositories list
type Catalog struct {
	Repositories []string `json:"repositories"`
}

// fullCatalog returns the full catalog from the registry server
func fullCatalog() (*Catalog, error) {
	cat := &Catalog{}
	var e error
	c := &http.Client{}
	var rh string
	rh = os.Getenv("REGISTRY_SCHEMA") + "://" + os.Getenv("REGISTRY_HOST") + ":" + os.Getenv("REGISTRY_PORT")
	r, re := http.NewRequest("GET", rh+"/"+os.Getenv("REGISTRY_VERSION")+"/_catalog", nil)
	if re != nil {
		return cat, re
	}
	r.SetBasicAuth(os.Getenv("REGISTRY_ADMIN_USER"), os.Getenv("REGISTRY_ADMIN_PASS"))
	res, err := c.Do(r)
	if err != nil {
		return cat, err
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return cat, berr
	}
	jerr := json.Unmarshal(bd, &cat)
	if jerr != nil {
		return cat, jerr
	}
	return cat, e
}

// Handler returns list of repositories user can access
func Handler(w http.ResponseWriter, r *http.Request) {
	_, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Docker Registry"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized) + "\n"))
		return
	}
	cat, err := fullCatalog()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, uerr := users.GetCurrent(r)
	if uerr != nil {
		http.Error(w, uerr.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
	if u.Admin {
		jd, jerr := json.Marshal(&cat)
		if jerr != nil {
			http.Error(w, jerr.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, string(jd))
		return
	}
	var urs []string
	repos := cat.Repositories
	for _, repo := range repos {
		for _, ns := range u.Namespaces {
			rm := regexp.MustCompile("^" + ns + "/")
			if repo != "" && ns != "" && rm.MatchString(repo) {
				urs = append(urs, repo)
			}
		}
	}
	uc := Catalog{
		Repositories: urs,
	}
	jd, jerr := json.Marshal(&uc)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, string(jd))
}
