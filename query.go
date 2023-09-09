package main

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

//go:embed html/query.tmpl
var queryHTML string

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

var queryTmpl *template.Template

func init() {
	queryTmpl = template.Must(template.New("query").
		Funcs(template.FuncMap{
			"sec":        func(millis int64) string { return fmt.Sprintf("%d.%03d", millis/1000, millis%1000) },
			"styleAttr":  styleAttr,
			"styleValue": styleValue,
			"tdValue":    tdValue,
		}).
		Parse(queryHTML))
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

var (
	// SQL definitions and displays are managed separately. hide and order directly in the query.tmpl
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
