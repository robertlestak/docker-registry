package catalog

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func admins() ([]string, error) {
	var as []string
	var e error
	af, err := ioutil.ReadFile(os.Getenv("RM_ADMIN_PASS_FILE"))
	if err != nil {
		return as, err
	}
	afs := string(af)
	afss := strings.SplitN(afs, "\n", -1)
	for _, aup := range afss {
		up := strings.SplitN(aup, ":", -1)
		if len(up) > 0 {
			as = append(as, up[0])
		}
	}
	return as, e
}

func isAdmin(u string) bool {
	var ia bool
	as, err := admins()
	if err != nil {
		return false
	}
	for _, au := range as {
		if u == au {
			return true
		}
	}
	return ia
}

func userNamespaces(u string) ([]string, error) {
	var ns []string
	var e error
	allUsers, err := ioutil.ReadDir(os.Getenv("RM_NAMESPACE_DIR"))
	if err != nil {
		return ns, err
	}
	for _, un := range allUsers {
		if u == un.Name() {
			uns, une := ioutil.ReadFile(path.Join(os.Getenv("RM_NAMESPACE_DIR"), un.Name()))
			if une != nil {
				return ns, une
			}
			ns = strings.SplitN(string(uns), "\n", -1)
		}
	}
	return ns, e
}
