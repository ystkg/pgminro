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

	SslDisable    string
	SslAllow      string
	SslPrefer     string
	SslRequire    string
	SslVerifyCa   string
	SslVerifyFull string

	ReadOnlyOn  string
	ReadOnlyOff string

	ErrorMessage string
}

var connectTmpl *template.Template

func init() {
	osUser := "OS user"
	if user, err := user.Current(); err == nil {
		osUser = user.Username
	}
	connectTmpl = template.Must(template.New("connect").
		Funcs(template.FuncMap{"valueAttr": valueAttr}).
		Parse(strings.Replace(connectHTML, "osUser", osUser, 1)))
}

func valueAttr(val string) template.HTMLAttr {
	s := ""
	if val != "" {
		s = fmt.Sprintf(" value='%s'", html.EscapeString(val))
	}
	return template.HTMLAttr(s)
}

func newConnectInfo(driver string, form *ConnectForm, err error, opts map[string]string) *ConnectInfo {
	connectInfo := &ConnectInfo{}
	if driver == "pgx" {
		connectInfo.Pgx = "checked"
	} else if driver == "pq" {
		connectInfo.Pq = "checked"
	} else if defaultPgx {
		connectInfo.Pgx = "checked"
	} else {
		connectInfo.Pq = "checked"
	}
	if form != nil {
		connectInfo.ConnectForm = *form
	}
	if err != nil {
		connectInfo.ErrorMessage = err.Error()
	}
	if opts == nil {
		connectInfo.SslDisable = "checked"
		connectInfo.ReadOnlyOn = "checked"
	} else {
		switch opts["sslmode"] {
		case "allow":
			connectInfo.SslAllow = "checked"
		case "prefer":
			connectInfo.SslPrefer = "checked"
		case "require":
			connectInfo.SslRequire = "checked"
		case "verify-ca":
			connectInfo.SslVerifyCa = "checked"
		case "verify-full":
			connectInfo.SslVerifyFull = "checked"
		default:
			connectInfo.SslDisable = "checked"
		}
		if opts["readonly"] == "off" {
			connectInfo.ReadOnlyOff = "checked"
		} else {
			connectInfo.ReadOnlyOn = "checked"
		}
	}
	return connectInfo
}
