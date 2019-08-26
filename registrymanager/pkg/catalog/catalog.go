package catalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/umg/docker-registry-manager/pkg/users"
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
	User *users.User
}

func catalogReq(req *Request) (*Catalog, *Request, http.Header, error) {
	c := &http.Client{}
	cat := &Catalog{}
	var rurl string
	requrl, rerr := catalogURL(req)
	if rerr != nil {
		return cat, req, nil, rerr
	}
	if req.User.Admin {
		rurl = requrl
	} else {
		ru, _ := url.Parse(requrl)
		params := url.Values{}
		params.Set("n", strconv.Itoa(100))
		params.Set("last", req.Last)
		ru.RawQuery = params.Encode()
		rurl = ru.String()
	}

	r, re := http.NewRequest("GET", rurl, nil)
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
	pn := req.Num
	if res.Header.Get("link") != "" {
		rq, rerr := parseNextLink(res.Header.Get("link"))
		if rerr != nil {
			return cat, rq, res.Header, rerr
		}
		req = rq
	} else {
		req.Last = ""
	}
	req.Num = pn
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
	req.User = u
	uc := buildUserCatalog(w, h, cat, req, u)
	jd, jerr := json.Marshal(&uc)
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, string(jd))
}
