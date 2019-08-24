package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// BasicAuthRealm is the string name of the realm
const BasicAuthRealm string = "Docker Registry"

// Catalog contains the repositories list
type Catalog struct {
	Repositories []string `json:"repositories"`
}

// Request contains a request to the registry
type Request struct {
	Num  int
	Last string
}

func catalogReq(req *Request) (*Catalog, *Request, http.Header, error) {
	c := &http.Client{}
	cat := &Catalog{}
	requrl, rerr := catalogURL(req)
	if rerr != nil {
		return cat, req, nil, rerr
	}
	r, re := http.NewRequest("GET", requrl, nil)
	if re != nil {
		return cat, req, nil, re
	}
	r.SetBasicAuth(os.Getenv("REGISTRY_ADMIN_USER"), os.Getenv("REGISTRY_ADMIN_PASS"))
	res, err := c.Do(r)
	if err != nil {
		return cat, req, res.Header, err
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return cat, req, res.Header, berr
	}
	if res.Header.Get("link") != "" {
		rq, rerr := parseNextLink(res.Header.Get("link"))
		if rerr != nil {
			return cat, rq, res.Header, rerr
		}
		req = rq
	} else {
		req.Last = ""
	}
	jerr := json.Unmarshal(bd, &cat)
	if jerr != nil {
		return cat, req, res.Header, jerr
	}
	return cat, req, res.Header, nil
}

// Handler returns list of repositories user can access
func Handler(w http.ResponseWriter, r *http.Request) {
	_, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, BasicAuthRealm))
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized) + "\n"))
		return
	}
	cat, u, req, h, err := userCatalogHandler(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.Admin {
		setHeaders(w, h, u)
		jd, jerr := json.Marshal(&cat)
		if jerr != nil {
			http.Error(w, jerr.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, string(jd))
		return
	}
	uc := buildUserCatalog(w, h, cat, req, u)
	jd, jerr := json.Marshal(&uc)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, string(jd))
}
