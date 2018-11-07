package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

type Catalog struct {
	Repositories []string `json:"repositories"`
}

func fullCatalog() (*Catalog, error) {
	cat := &Catalog{}
	var e error
	c := &http.Client{}
	var rh string
	if os.Getenv("REGISTRY_HOST") != "" {
		rh = os.Getenv("REGISTRY_HOST")
	} else {
		rh = os.Getenv("PROXY_DOMAIN")
	}
	r, re := http.NewRequest("GET", rh+"/v2/_catalog", nil)
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

func outputCatalog(w http.ResponseWriter, c *Catalog) {
	fmt.Fprint(w, c)
}

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	cat, err := fullCatalog()
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	u := r.Header.Get("X-Auth-User")
	if isAdmin(u) {
		jd, jerr := json.Marshal(&cat)
		if jerr != nil {
			fmt.Fprintf(w, jerr.Error())
			return
		}
		fmt.Fprintf(w, string(jd))
		return
	}
	uns, nse := userNamespaces(u)
	if nse != nil {
		fmt.Println(nse)
		return
	}
	fc, fce := fullCatalog()
	if fce != nil {
		fmt.Fprintf(w, fce.Error())
		return
	}
	var urs []string
	repos := fc.Repositories
	for _, repo := range repos {
		for _, ns := range uns {
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
		fmt.Fprintf(w, jerr.Error())
		return
	}
	fmt.Fprintf(w, string(jd))
}
