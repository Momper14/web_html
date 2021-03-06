package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/Momper14/web/templates"
	"github.com/Momper14/weblib/client/karteikaesten"
	"github.com/Momper14/weblib/client/users"
	"github.com/fatih/structs"
)

func customExecuteTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {

	type (
		Sidemenu struct {
			Kasten      int
			MeineKasten int
		}

		Navbar struct {
			Name string
			Bild string
		}
	)

	var (
		navbar   string
		sidemenu string
		dataTmp  map[string]interface{}
		sm       Sidemenu
		nb       Navbar
		err      error
	)

	if test := IstEingeloggt(w, r); test {
		navbar = templates.NavbarLogin
		sidemenu = templates.SidemenuLogin

		if sm.MeineKasten, err = karteikaesten.New().AnzahlKaestenUser(GetUser(w, r)); err != nil {
			internalError(err, w)
		}

		nb.Name = GetUser(w, r)
		user, err := users.New().UserByID(nb.Name)
		errF(err, w)

		nb.Bild, err = imageLastmod(user.Bild)
		errF(err, w)
	} else {
		navbar = templates.NavbarNoLogin
		sidemenu = templates.SidemenuNoLogin
	}

	if structs.IsStruct(data) {
		dataTmp = structs.Map(data)
	} else {
		var ok bool
		dataTmp, ok = data.(map[string]interface{})
		if !ok {
			dataTmp = make(map[string]interface{})
		}
	}

	if sm.Kasten, err = karteikaesten.New().AnzahlOeffentlicherKaesten(); err != nil {
		internalError(err, w)
	}

	dataTmp["Sidemenu"] = sm
	dataTmp["Navbar"] = nb

	t, err := template.ParseFiles(templateName, navbar, sidemenu)
	errF(err, w)

	errF(t.Execute(w, dataTmp), w)

}

func internalError(err error, w http.ResponseWriter) {
	http.Error(w, "500 - Something bad happened!", http.StatusInternalServerError)

	panic(err)
}

func recoverInternalError() {
	if r := recover(); r != nil {
		fmt.Println(r)
	}
}

func forbidden(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func errF(err error, w http.ResponseWriter) {
	if err != nil {
		internalError(err, w)
	}
}

// imageLastmod returns the file with ?lastmod=...
func imageLastmod(file string) (string, error) {

	tmp, err := os.Stat(file[1:])

	if err != nil {
		return "", err
	}

	file = fmt.Sprintf("%s?lastmod=%d", file, tmp.ModTime().Unix())

	return file, nil
}
