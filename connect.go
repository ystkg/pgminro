package main

import (
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"os/user"
	"strings"
)

//go:embed html/connect.tmpl
var connectHTML string

type ConnectForm struct {
	Host     string
	Port     string
	Database string
	User     string
}

type ConnectInfo struct {
	ConnectForm

	Pq  string
	Pgx string

	ErrorMessage string
}

var connectTmpl *template.Template

func init() {
	connStrFmt := formatDSN(ConnectForm{"<i><b>host</b></i>", "<i><b>port</b></i>", "<i><b>database</b></i>", "<i><b>user</b></i>"}, "<i><b>password</b></i>")
	osUser := "OS user"
	if user, err := user.Current(); err == nil {
		osUser = user.Username
	}
	connectTmpl = template.Must(template.New("connect").
		Funcs(template.FuncMap{"valueAttr": valueAttr}).
		Parse(strings.Replace(strings.Replace(connectHTML, "connStrFmt", connStrFmt, 1), "osUser", osUser, 1)))
}

func valueAttr(val string) template.HTMLAttr {
	s := ""
	if val != "" {
		s = fmt.Sprintf(" value='%s'", html.EscapeString(val))
	}
	return template.HTMLAttr(s)
}
