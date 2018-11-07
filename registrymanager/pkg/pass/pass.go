package pass

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func setPass(u string, p string, f string) error {
	var e error
	cmd := exec.Command("htpasswd", "-nb", u, p)
	up, oerr := cmd.Output()
	if oerr != nil {
		return oerr
	}
	ups := strings.Replace(string(up), "\n", "", -1)
	sscr := "s|" + u + ":.*$|" + ups + "|g"
	sed := exec.Command("sed", sscr, f)
	sedout, soerr := sed.Output()
	if soerr != nil {
		return soerr
	}
	werr := ioutil.WriteFile(f, sedout, 0644)
	if werr != nil {
		return werr
	}
	return e
}

func fileContainsUser(f string, u string) (bool, error) {
	var ex bool
	var e error
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return ex, e
	}
	s := string(b)
	if strings.Contains(s, u+":") {
		return true, e
	}
	return ex, e
}

func usernameExists(u string) (string, error) {
	var exf string
	var e error
	ac, aerr := fileContainsUser(os.Getenv("RM_ADMIN_PASS_FILE"), u)
	if aerr != nil {
		return exf, aerr
	}
	if ac {
		exf = os.Getenv("RM_ADMIN_PASS_FILE")
	} else {
		uc, uerr := fileContainsUser(os.Getenv("RM_USER_PASS_FILE"), u)
		if uerr != nil {
			return exf, uerr
		}
		if uc {
			exf = os.Getenv("RM_USER_PASS_FILE")
		}
	}
	return exf, e
}

func rebuildConf() error {
	var e error
	cmd := exec.Command("make", "reload")
	cmd.Dir = os.Getenv("RM_PATH_TO_MAKEFILE")
	_, oerr := cmd.Output()
	if oerr != nil {
		return oerr
	}
	return e
}

func PasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
	u := r.Header.Get("X-Auth-User")
	p := r.FormValue("password")
	if u == "" || p == "" {
		fmt.Fprint(w, "username and password required")
		return
	}
	uexf, uerr := usernameExists(u)
	if uerr != nil {
		fmt.Fprint(w, uerr.Error())
		return
	} else if uexf == "" {
		fmt.Fprint(w, "Username does not exist")
		return
	}
	if err := setPass(u, p, uexf); err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	go rebuildConf()
	fmt.Fprint(w, "Password changed")
}
