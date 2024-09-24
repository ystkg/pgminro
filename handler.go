package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type session struct {
	lastAccessAt time.Time

	conn *connection

	paramInRedirect string
}

var (
	sessionStore = make(map[string]*session)
	mu           sync.Mutex
	once         sync.Once
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var sessionId, param string
	var session *session
	if c, err := r.Cookie("PGSESSIONID"); err == nil {
		sessionId = c.Value
		if session = sessionStore[sessionId]; session != nil {
			session.lastAccessAt = time.Now()
			param, session.paramInRedirect = session.paramInRedirect, ""
		}
	}
	if session == nil {
		id, s, err := generateSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "PGSESSIONID",
			Value:    id,
			Path:     "/",
			Secure:   useHttpsCertainty,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		session = s
	}

	action := r.PostFormValue("action")

	var connectInfo *ConnectInfo
	if session.conn == nil {
		if action == "connect" {
			form := ConnectForm{
				Host:     r.PostFormValue("host"),
				Port:     r.PostFormValue("port"),
				Database: r.PostFormValue("database"),
				User:     r.PostFormValue("user"),
			}
			driver := r.PostFormValue("driver")
			if conn, err := openDB(driver, form, r.PostFormValue("password")); err == nil {
				session.conn = conn
			} else {
				connectInfo = &ConnectInfo{form, "", "", err.Error()}
				if driver == "pgx" {
					connectInfo.Pgx = "checked"
				} else {
					connectInfo.Pq = "checked"
				}
			}
		} else {
			connectInfo = &ConnectInfo{Pq: "checked"}
		}
	} else if action == "disconnect" {
		invalidateSession(sessionId)
		connectInfo = &ConnectInfo{Pq: "checked"}
	}

	if connectInfo != nil { // unconnected
		setHeaders(w)
		if err := connectTmpl.Execute(w, connectInfo); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
		}
		return
	}

	if table := r.FormValue("table"); table != "" && r.Method == "GET" && len(r.Form) == 1 { // by hyperlink
		session.paramInRedirect = table
		http.Redirect(w, r, "/", http.StatusFound) // for remove querystring in url
		return
	}

	var sql, sqlkey string
	switch action {
	case "execute":
		sql = r.PostFormValue("sql")
	case "sqldef":
		sqlkey = r.PostFormValue("sqlkey")
		sql = sqlMapping[sqlkey]
	case "":
		if param != "" && r.Method == "GET" && len(r.Form) == 0 { // by redirect
			sql = "SELECT * FROM " + param
		}
	}

	queryData := &QueryData{
		QueryForm: QueryForm{sql},
		ConnStr:   session.conn.connStr,
	}

	if sql != "" {
		if rs, err := session.conn.Query(sql); err == nil {
			queryData.Count = len(rs.Rows)
			queryData.ResultSet = rs
			queryData.IsExplain = isExplain(sql)
			queryData.HyperlinkIndex = hyperlinkIndex(rs.Names, sqlkey)
		} else {
			queryData.ErrorMessage = err.Error()
		}
	}

	setHeaders(w)
	if err := queryTmpl.Execute(w, queryData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func setHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Cache-Control", "no-cache, no-store, max-age=0, private, must-revalidate")
	header.Set("Pragma", "no-cache")
	header.Set("Expires", "0")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
	header.Set("Content-Security-Policy", "default-uri 'none'")
	if useHttpsCertainty {
		header.Set("Strict-Transport-Security", "max-age=31536000")
	}
}

const EXPLAIN = "EXPLAIN "

func isExplain(sql string) bool {
	t := strings.TrimSpace(sql)
	if len(t) < len(EXPLAIN) {
		return false
	}
	return strings.ToUpper(t[:len(EXPLAIN)]) == EXPLAIN
}

func hyperlinkIndex(names []string, sqlkey string) int {
	if sqlkey != "" {
		hyperlink := hyperlinkMapping[sqlkey]
		for i, v := range names {
			if v == hyperlink {
				return i
			}
		}
	}
	return -1
}

func generateSession() (string, *session, error) {
	mu.Lock()
	defer mu.Unlock()

	once.Do(func() { go sessionLifecycleTimer(context.Background(), sessionTimeoutMin*time.Minute) })

	bytes := make([]byte, 20)
	for {
		n, err := rand.Reader.Read(bytes)
		if err != nil {
			return "", nil, err
		}
		if n != len(bytes) {
			continue
		}
		candi := hex.EncodeToString(bytes)
		if _, already := sessionStore[candi]; already {
			continue
		}
		session := &session{lastAccessAt: time.Now()}
		sessionStore[candi] = session
		return candi, session, nil
	}
}

func invalidateSession(sessionId string) {
	session := sessionStore[sessionId]
	if session == nil {
		return
	}
	var conn *connection
	conn, session.conn = session.conn, nil
	if conn == nil {
		return
	}
	conn.Close()
	sessionStore[sessionId] = nil // no delete
}

func sessionLifecycleTimer(ctx context.Context, sessionTimeout time.Duration) {
	sleep := sessionTimeout
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
			sleep = sessionTimeout

			mu.Lock()
			now := time.Now()
			for k, v := range sessionStore {
				if v == nil {
					continue
				}
				if life := sessionTimeout - now.Sub(v.lastAccessAt); 0 < life {
					sleep = min(sleep, life)
					continue
				}
				invalidateSession(k)
			}
			mu.Unlock()

			sleep = max(sleep, time.Second)
		}
	}
}
