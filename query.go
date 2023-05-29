package main

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

const queryHTML = `
<!DOCTYPE html>
<html lang='ja'>
<head>
<meta charset='utf-8'>
</head>
<body>
<header>
<form action='/' method='POST'>
<table style='width: 100%;'>
<tr>
<td>{{.ConnStr}}</td>
<td style='text-align: right;'><input type='submit' name='action' value='disconnect'></td>
</tr>
</table>
</form>
</header>
<nav>
<form action='/' method='POST'>
<input type='hidden' name='action' value='sqldef'>
<table>
<tr>
<td><input type='submit' name='sqlkey' value='SYSTEM_INFO'></td>
<td><input type='submit' name='sqlkey' value='SETTINGS'></td>
<td><input type='submit' name='sqlkey' value='FILE_SETTINGS'></td>
<td><input type='submit' name='sqlkey' value='HBA'></td>
<td><input type='submit' name='sqlkey' value='PG_CATALOG'></td>
<td><input type='submit' name='sqlkey' value='INFORMATION_SCHEMA'></td>
<td><input type='submit' name='sqlkey' value='USER'></td>
<td><input type='submit' name='sqlkey' value='ROLE'></td>
<td><input type='submit' name='sqlkey' value='DB'></td>
<td><input type='submit' name='sqlkey' value='TABLE'></td>
</tr>
</table>
</form>
</nav>
<main>
<form action='/' method='POST'>
<textarea name='sql' style='resize: vertical; width: 100%; height: 300px;'>{{.SQL}}</textarea>
<input type='submit' name='action' value='execute'>
</form>
<br>
{{$h := .HyperlinkIndex}}
{{if .ErrorMessage}}
<pre>
{{.ErrorMessage}}
</pre>
{{else if .IsExplain}}
<pre>
{{range .Rows}}{{range .Values}}{{.}}
{{end}}{{end}}
</pre>
{{else if .ResultSet}}
	{{if .Over}}<span style='color: red;'>over {{.Count}} rows! ({{.Time | sec}}sec)</span>{{else}}<span>{{.Count}} rows ({{.Time | sec}}sec)</span>{{end}}
	<table style='border-collapse: collapse;'>
	<thead>
	<tr style='background-color: darkgray;'>
	{{range .Names}}
		<th style='border: 1px solid gray;'>{{.}}</th>
	{{end}}
	</tr>
	<tr style='background-color: darkgray; text-align: center;'>
	{{range .Types}}
		<td style='border: 1px solid gray;'>({{.}})</td>
	{{end}}
	</tr>
	</thead>
	<tbody>
	{{range $i, $_ := .Rows}}
		<tr{{$i | styleAttr}}>
		{{range $j, $_ := .Values}}
			<td style='{{. | styleValue}}'>{{. | tdValue $j $h}}</td>
		{{end}}
		</tr>
	{{end}}
{{end}}
</tbody>
</table>
</main>
</body>
</html>
`

type QueryForm struct {
	SQL string
}

type QueryData struct {
	QueryForm

	ConnStr string

	Count int
	*ResultSet

	IsExplain      bool
	HyperlinkIndex int

	ErrorMessage string
}

var (
	// SQL definitions and displays are managed separately. hide and order directly in the queryHTML.
	sqlMapping = map[string]string{
		"SYSTEM_INFO":        "SELECT current_timestamp, inet_server_addr(), inet_server_port(), current_database(), current_schema(), version(), pg_postmaster_start_time(), pg_conf_load_time(), txid_current_snapshot()",
		"SETTINGS":           "SELECT category, name, setting, unit, source, short_desc FROM pg_settings ORDER BY category, name",
		"FILE_SETTINGS":      "SELECT seqno, sourcefile, sourceline, name, setting, applied, error FROM pg_file_settings ORDER BY seqno",
		"HBA":                "SELECT * FROM pg_hba_file_rules ORDER BY line_number",
		"PG_CATALOG":         "SELECT name, type FROM (SELECT schemaname, tablename AS name, 'table' AS type FROM pg_tables UNION SELECT schemaname, viewname AS name, 'view' AS type FROM pg_views) as A WHERE schemaname = 'pg_catalog' ORDER BY name",
		"INFORMATION_SCHEMA": "SELECT name, type, schemaname || '.' || name AS fullname FROM (SELECT schemaname, tablename AS name, 'table' AS type FROM pg_tables UNION SELECT schemaname, viewname AS name, 'view' AS type FROM pg_views) as A WHERE schemaname = 'information_schema' ORDER BY name",
		"USER":               "SELECT * FROM pg_user ORDER BY usesysid",
		"ROLE":               "SELECT * FROM pg_roles ORDER BY oid",
		"DB":                 "SELECT * FROM pg_database ORDER BY oid",
		"TABLE":              "SELECT schemaname, tablename, schemaname || '.' || tablename AS fullname FROM pg_tables WHERE schemaname <> 'pg_catalog' AND schemaname <> 'information_schema' ORDER BY schemaname, tablename",
	}

	hyperlinkMapping = map[string]string{
		"PG_CATALOG":         "name",
		"INFORMATION_SCHEMA": "fullname",
		"TABLE":              "fullname",
	}
)

var queryTmpl *template.Template

func init() {
	queryTmpl = template.Must(template.New("query").
		Funcs(template.FuncMap{
			"sec":        func(millis int64) string { return fmt.Sprintf("%d.%03d", millis/1000, millis%1000) },
			"styleAttr":  styleAttr,
			"styleValue": styleValue,
			"tdValue":    tdValue,
		}).
		Parse(queryHTML[1:]))
}

func styleAttr(i int) template.HTMLAttr {
	style := ""
	if i%2 != 0 {
		style = " style='background-color: lightcyan;'"
	}
	return template.HTMLAttr(style)
}

func styleValue(v any) template.CSS {
	style := "white-space: pre-wrap; border: 1px solid gray;"
	switch p := v.(type) {
	case *bool:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		}
	case *int64:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		} else {
			style += " text-align: right;"
		}
	case *uint64:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		} else {
			style += " text-align: right;"
		}
	case *float32:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		} else {
			style += " text-align: right;"
		}
	case *float64:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		} else {
			style += " text-align: right;"
		}
	case *string:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		}
	case *DateTime:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		}
	case *ByteArray:
		if p == nil {
			style += " text-align: center; color: lightgray;"
		}
	default:
		style += " text-align: center; color: dimgray;"
	}
	return template.CSS(style)
}

func tdValue(i, hyperlinkIndex int, v any) template.HTML {
	s := "(null)"
	switch p := v.(type) {
	case *bool:
		if p != nil {
			s = strconv.FormatBool(*p)
		}
	case *int64:
		if p != nil {
			s = strconv.FormatInt(*p, 10)
		}
	case *uint64:
		if p != nil {
			s = strconv.FormatUint(*p, 10)
		}
	case *float32:
		if p != nil {
			s = strconv.FormatFloat(float64(*p), 'f', -1, 32)
		}
	case *float64:
		if p != nil {
			s = strconv.FormatFloat(*p, 'f', -1, 64)
		}
	case *string:
		if p != nil {
			s = *p
			if i == hyperlinkIndex {
				table := s
				if strings.HasPrefix(table, "public.") {
					table = s[7:]
				}
				s = fmt.Sprintf("<a href='/?table=%s'>%s</a>", table, s)
			}
		}
	case *DateTime:
		if p.time != nil {
			switch p.databaseType {
			case "DATE":
				s = p.time.Format(time.DateOnly)
			default:
				s = p.time.Format(time.DateTime + ".000")
			}
		}
	case *ByteArray:
		if p.bytes != nil {
			switch p.databaseType {
			case "UUID":
				s = hex.EncodeToString(*p.bytes)
			default:
				s = string(*p.bytes)
			}
		}
	default:
		s = "(unknown)"
	}
	return template.HTML(s)
}
