package main

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/lib/pq"
)

type connection struct {
	db      *sql.DB
	connStr string
}

type ResultRow struct {
	Values []any
}

type ResultSet struct {
	Time  int64
	Over  bool
	Names []string
	Types []string
	Rows  []ResultRow
}

type DateTime struct {
	databaseType string
	time         *time.Time
}

type ByteArray struct {
	databaseType string
	bytes        *[]byte
}

func formatDSN(form ConnectForm, password string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", form.User, password, form.Host, form.Port, form.Database)
}

func openDB(form ConnectForm, password string) (*connection, error) {
	db, err := sql.Open("postgres", formatDSN(form, password))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(maxIdleTimeMin * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeoutSec*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	masked := ""
	if password != "" {
		masked = "********"
	}
	return &connection{db, formatDSN(form, masked)}, nil
}

func (conn *connection) Close() error {
	var d *sql.DB
	d, conn.db = conn.db, nil
	if d == nil {
		return nil
	}
	return d.Close()
}

func (conn *connection) Query(query string) (*ResultSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeoutSec*time.Second)
	defer cancel()

	tx, err := conn.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	start := time.Now()
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	cols, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	types := make([]string, len(cols))
	columns := make([]any, len(cols))
	for i, v := range cols {
		types[i] = v.DatabaseTypeName()
		columns[i] = alloc(v)
	}

	over := false
	resultRows := make([]ResultRow, 0)
	for i := 0; rows.Next(); i++ {
		if err = rows.Scan(columns...); err != nil {
			return nil, err
		}
		if maxRows <= i {
			over = true
			break
		}
		values := make([]any, len(columns))
		for i, v := range columns {
			values[i] = ptr(v, types[i])
		}
		resultRows = append(resultRows, ResultRow{values})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	rs := &ResultSet{
		Time:  time.Since(start).Milliseconds(),
		Over:  over,
		Names: names,
		Types: types,
		Rows:  resultRows,
	}

	return rs, nil
}

func alloc(v *sql.ColumnType) any {
	switch v.ScanType().Kind() {
	case reflect.Bool:
		return new(*bool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(*int64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(*uint64)
	case reflect.Float32:
		return new(*float32)
	case reflect.Float64:
		return new(*float64)
	case reflect.String:
		return new(*string)
	case reflect.Interface:
		switch v.DatabaseTypeName() {
		case "FLOAT4":
			return new(*float32)
		case "FLOAT8":
			return new(*float64)
		case "NAME", "CHAR":
			return new(*string)
		case "OID", "XID":
			return new(*uint64)
		}
	case reflect.Struct:
		if v.ScanType().String() == "time.Time" {
			return new(*time.Time)
		}
	}

	return new(*[]byte)
}

func ptr(v any, databaseTypeName string) any {
	switch p := v.(type) {
	case **bool:
		return *p
	case **int64:
		return *p
	case **uint64:
		return *p
	case **float32:
		return *p
	case **float64:
		return *p
	case **string:
		return *p
	case **time.Time:
		return &DateTime{
			databaseType: databaseTypeName,
			time:         *p,
		}
	case **[]byte:
		return &ByteArray{
			databaseType: databaseTypeName,
			bytes:        *p,
		}
	}
	return v
}
