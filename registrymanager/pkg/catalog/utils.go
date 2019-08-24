package catalog

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/umg/docker-registry-manager/pkg/users"
)

func catalogURL(req *Request) (string, error) {
	var rh string
	rh = os.Getenv("REGISTRY_SCHEMA") + "://" + os.Getenv("REGISTRY_HOST") + ":" + os.Getenv("REGISTRY_PORT")
	requrl, rerr := url.Parse(rh)
	if rerr != nil {
		return "", rerr
	}
	requrl.Path += "/" + os.Getenv("REGISTRY_VERSION") + "/_catalog"
	params := url.Values{}
	if req.Last != "" {
		params.Add("last", req.Last)
	}
	if req.Num > 0 {
		params.Add("n", strconv.Itoa(req.Num))
	}
	requrl.RawQuery = params.Encode()
	return requrl.String(), nil
}

func parseNextLink(h string) (*Request, error) {
	r := &Request{}
	l := strings.Split(h, ";")
	tl := strings.Replace(l[0], "<", "", 1)
	tl = strings.Replace(tl, ">", "", 1)
	qs := strings.Split(tl, "?")
	q, err := url.ParseQuery(qs[1])
	if err != nil {
		return r, err
	}
	var e error
	r.Last = q.Get("last")
	r.Num, e = strconv.Atoi(q.Get("n"))
	if e != nil {
		return r, e
	}
	return r, nil
}

// parseRequest parses the request details into a struct
func parseRequest(r *http.Request) (*Request, error) {
	r.ParseForm()
	req := &Request{
		Last: r.FormValue("last"),
	}
	if r.FormValue("n") != "" {
		var perr error
		req.Num, perr = strconv.Atoi(r.FormValue("n"))
		if perr != nil {
			return req, perr
		}
		if req.Num > 100 {
			req.Num = 100
		}
	}
	return req, nil
}

func userAccessRepo(u *users.User, repo string) bool {
	for _, ns := range u.Namespaces {
		rm := regexp.MustCompile("^" + ns + "/")
		if repo != "" && ns != "" && rm.MatchString(repo) {
			return true
		}
	}
	return false
}

func trimUserRepos(cat *Catalog, u *users.User) []string {
	var urs []string
	repos := cat.Repositories
	for _, repo := range repos {
		if userAccessRepo(u, repo) {
			urs = append(urs, repo)
		}
	}
	return urs
}

func shouldCheckNextPage(uc *Catalog, req *Request, u *users.User) bool {
	if req.Last == "" {
		return false
	} else if len(uc.Repositories) == 0 {
		return true
	} else if len(uc.Repositories) >= 100 {
		return false
	} else if len(uc.Repositories) < req.Num && userAccessRepo(u, uc.Repositories[len(uc.Repositories)-1]) {
		return true
	}
	return false
}

// buildUserCatalog iterates through the repos and ensures only the user's namespaces are displayed
func buildUserCatalog(w http.ResponseWriter, h http.Header, cat *Catalog, req *Request, u *users.User) *Catalog {
	urs := trimUserRepos(cat, u)
	uc := &Catalog{
		Repositories: urs,
	}
	var lastElem string
	for shouldCheckNextPage(uc, req, u) {
		cat, lreq, _, err := catalogReq(req)
		if err != nil {
			return uc
		}
		req = lreq
		lurs := trimUserRepos(cat, u)
		uc.Repositories = append(uc.Repositories, lurs...)
		if !userAccessRepo(u, lurs[len(lurs)-1]) {
			break
		} else {
			lastElem = lurs[len(lurs)-1]
		}
	}
	if len(uc.Repositories) >= 100 {
		h.Set("link", "</v2/_catalog?last="+lastElem+"&n=100>; rel=\"next\"")
	} else {
		h.Del("link")
	}
	setHeaders(w, h, u)
	return uc
}

// userCatalogHandler handles the api request for a user's catalog
func userCatalogHandler(w http.ResponseWriter, r *http.Request) (*Catalog, *users.User, *Request, http.Header, error) {
	req, rerr := parseRequest(r)
	if rerr != nil {
		return nil, nil, req, nil, rerr
	}
	cat, rreq, h, err := catalogReq(req)
	if err != nil {
		return cat, nil, rreq, h, err
	}
	u, uerr := users.GetCurrent(r)
	if uerr != nil {
		return cat, u, rreq, h, uerr
	}
	return cat, u, rreq, h, nil
}

// setHeaders sets the response headers for API calls
func setHeaders(w http.ResponseWriter, h http.Header, u *users.User) {
	w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
	for k, v := range h {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		w.Header().Set(k, strings.Join(v, ","))
	}
}
