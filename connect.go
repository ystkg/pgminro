package main

import (
	"fmt"
	"html"
	"html/template"
	"os/user"
	"strings"
)

const connectHTML = `
<!DOCTYPE html>
<html lang='ja'>
<head>
<meta charset='utf-8'>
</head>
<body>
<table style='width: 100%;'>
<tr>
<td>connStrFmt</td>
</tr>
</table>
<br>
<form action='/' method='POST'>
<table>
<tr><th style='text-align: left;'>host</th><td style='text-align: left;'><input type='text' size='35' maxlength='150' name='host'{{.Host | valueAttr}}></td><td>(default:localhost)</td></tr>
<tr><th style='text-align: left;'>port</th><td style='text-align: left;'><input type='text' size='6' maxlength='5' name='port'{{.Port | valueAttr}}></td><td>(default:5432)</td></tr>
<tr><th style='text-align: left;'>database</th><td style='text-align: left;'><input type='text' size='35' maxlength='50' name='database'{{.Database | valueAttr}}></td><td>(default:user name)</td></tr>
<tr><th style='text-align: left;'>user</th><td style='text-align: left;'><input type='text' size='35' maxlength='50' name='user'{{.User | valueAttr}}></td><td>(default:osUser)</td></tr>
<tr><th style='text-align: left;'>password</th><td style='text-align: left;'><input type='password' size='35' maxlength='100' name='password'></td><td></td></tr>
<tr><td></td><td><input type='submit' name='action' value='connect'></td><td></td></tr>
</table>
</form>
{{if ne .ErrorMessage ""}}
<pre>
{{.ErrorMessage}}
</pre>
{{end}}
</body>
</html>
`

type ConnectForm struct {
	Host     string
	Port     string
	Database string
	User     string
}

type ConnectInfo struct {
	ConnectForm

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
		Parse(strings.Replace(strings.Replace(connectHTML[1:], "connStrFmt", connStrFmt, 1), "osUser", osUser, 1)))
}

func valueAttr(val string) template.HTMLAttr {
	s := ""
	if val != "" {
		s = fmt.Sprintf(" value='%s'", html.EscapeString(val))
	}
	return template.HTMLAttr(s)
}
