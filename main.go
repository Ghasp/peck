package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	var err error
	CONNECTION, err = connectDB(SERVER, DATABASE, USERNAME, PASSWORD)
	if err != nil {
		fmt.Println("OOP!")
		panic(err)
	}
	defer CONNECTION.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")

	r.HandleFunc("/console", consoleHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/query", queryHandler).Methods("POST").Headers("Content-Type", "application/sql")
	r.HandleFunc("/modify", modifyHandler).Methods("POST").Headers("Content-Type", "application/sql")

	r.HandleFunc("/{schema}/{table}", getRowHandler).Methods("GET")
	r.HandleFunc("/{schema}/{table}/{key}", getRowHandler).Methods("GET")
	r.HandleFunc("/{schema}/{table}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}/{key}", putRowHandler).Methods("PUT").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}", postRowHandler).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{schema}/{table}", deleteRowHandler).Methods("DELETE")

	http.ListenAndServe("localhost:8080", r)
}

// TOKEN global variable
var TOKEN string = ""

// SERVER global variable
var SERVER string = ""

// DATABASE global variable
var DATABASE string = ""

// USERNAME global variable
var USERNAME string = ""

// PASSWORD global variable
var PASSWORD string = ""

// CONNECTION global variable
var CONNECTION *sql.DB

// PRIMARYKEY global variable
var PRIMARYKEY string = ""

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello")
}

func consoleHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	stmt, ok := getRequestBody(r)
	if !ok || len(stmt) == 0 {
		http.Error(w, `No sql statement provided. Please provide a sql statment in the http request body`, http.StatusBadRequest)
		return
	}

	conn, err := connectDB(SERVER, DATABASE, USERNAME, PASSWORD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	results, err := unsafeSQL(conn, string(stmt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	stmt, ok := getRequestBody(r)
	if !ok || len(stmt) == 0 {
		http.Error(w, `No sql statement provided. Please provide a sql statment in the http request body`, http.StatusBadRequest)
		return
	}

	data, err := getDataJSON(CONNECTION, string(stmt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, data)
}

func modifyHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	stmt, ok := getRequestBody(r)
	if !ok || len(stmt) == 0 {
		http.Error(w, `No sql statement provided. Please provide a sql statment in the http request body`, http.StatusBadRequest)
		return
	}

	err := setData(CONNECTION, string(stmt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func getRowHandler(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)
	schema := url["schema"]
	table := url["table"]

	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	where := make(map[string]string, 0)
	if key, ok := url["key"]; ok {
		where[PRIMARYKEY] = key
	} else {
		for k, v := range r.URL.Query() {
			where[k] = v[0]
		}
	}
	if len(where) < 1 {
		http.Error(w, `row index not specified`, http.StatusBadRequest)
		return
	}

	data, err := getRowJSON(CONNECTION, toSelect(DATABASE, schema, table, nil, where, true))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, data)

}

func putRowHandler(w http.ResponseWriter, r *http.Request) {
	url := mux.Vars(r)
	schema := url["schema"]
	table := url["table"]

	token, ok := parseToken(r)
	if !ok {
		http.Error(w, `token in the authorization header was not found.`, http.StatusBadRequest)
		return
	}
	if token != TOKEN {
		fmt.Println(token)
		requestToken(w)
		return
	}

	set, err := parseJSONRowRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	where := make(map[string]string, 0)
	if key, ok := url["key"]; ok {
		where[PRIMARYKEY] = key
	} else {
		for k, v := range r.URL.Query() {
			where[k] = v[0]
		}
	}
	if len(where) < 1 {
		http.Error(w, `row index not specified`, http.StatusBadRequest)
		return
	}

	err = setData(CONNECTION, toUpsert(DATABASE, schema, table, set, where))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
func postRowHandler(w http.ResponseWriter, r *http.Request)   {}
func deleteRowHandler(w http.ResponseWriter, r *http.Request) {}

func parseBasicAuth(r *http.Request) (string, string, bool) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		return "", "", false
	}

	payload, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		return "", "", false
	}

	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 || len(pair[0]) < 1 || len(pair[1]) < 1 {
		return "", "", false
	}
	return pair[0], pair[1], true
}

func parseToken(r *http.Request) (string, bool) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Bearer" {
		return "", false
	}

	if auth[1] == "" {
		return "", false
	}

	return auth[1], true
}

func requestCredentials(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Basic realm=\"SQL Credentials\"")
	http.Error(w, "", 401)
}

func requestToken(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", "Bearer realm=\"Token\"")
	http.Error(w, "", 401)
}

func getRequestBody(r *http.Request) ([]byte, bool) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}
	return b, true
}

func getData(db *sql.DB, sql string) ([]map[string]interface{}, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	pointers := newColumnPointers(columnTypes)

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		result := make(map[string]interface{}, 0)
		for i, column := range columnTypes {
			result[column.Name()] = pointers[i]
		}
		results = append(results, result)
	}

	if rows.Err() != nil {
		return nil, err
	}

	//query should never modify data.
	if err := tx.Rollback(); err != nil {
		return nil, err
	}

	return results, nil
}

func unsafeSQL(db *sql.DB, sql string) ([]map[string]interface{}, error) {

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	pointers := newColumnPointers(columnTypes)

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		result := make(map[string]interface{}, 0)
		for i, column := range columnTypes {
			result[column.Name()] = pointers[i]
		}
		results = append(results, result)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return results, nil
}

// Scan() can dynamically decode postgres data into some
// proveded types. this block sets up an array of destination pointers
// with a type that will be compatable with the postgres type.
// if the column is nullable, or the driver does not support this feature,
// the type will be a pointer.
func newColumnPointers(columnTypes []*sql.ColumnType) []interface{} {
	pointers := make([]interface{}, len(columnTypes))
	for i, column := range columnTypes {
		if nullable, ok := column.Nullable(); nullable || !ok {
			switch column.DatabaseTypeName() {
			case "BOOL":
				var t *bool
				pointers[i] = &t
			case "SMALLINT", "INTEGER", "BIGINT":
				var t *int64
				pointers[i] = &t
			case "SMALLSERIAL", "SERIAL", "BIGSERIAL":
				var t *uint64
				pointers[i] = &t
			case "DECIMAL", "NUMERIC", "REAL", "DOUBLE PRECISION", "FLOAT":
				var t *float64
				pointers[i] = &t
			case "TEXT", "CHARACTER", "CHARACTER VARYING":
				var t *string
				pointers[i] = &t
			case "TIME", "DATE", "TIMESTAMPT", "TIMESTAMPTZ":
				var t *time.Time
				pointers[i] = &t
			case "UUID":
				var t *uuid.UUID
				pointers[i] = &t
			case "BYTEA":
				var t *[]byte
				pointers[i] = &t
			default:
				var t interface{}
				pointers[i] = &t
			}
		} else {
			switch column.DatabaseTypeName() {
			case "BOOL":
				var t bool
				pointers[i] = &t
			case "SMALLINT", "INTEGER", "BIGINT":
				var t int64
				pointers[i] = &t
			case "SMALLSERIAL", "SERIAL", "BIGSERIAL":
				var t uint64
				pointers[i] = &t
			case "DECIMAL", "NUMERIC", "REAL", "DOUBLE PRECISION", "FLOAT":
				var t float64
				pointers[i] = &t
			case "TEXT", "CHARACTER", "CHARACTER VARYING":
				var t string
				pointers[i] = &t
			case "TIME", "DATE", "TIMESTAMPT", "TIMESTAMPTZ":
				var t time.Time
				pointers[i] = &t
			case "UUID":
				var t uuid.UUID
				pointers[i] = &t
			case "BYTEA":
				var t []byte
				pointers[i] = &t
			default:
				var t interface{}
				pointers[i] = &t
			}
		}
	}
	return pointers
}

func getDataJSON(db *sql.DB, sql string) (string, error) {
	results, err := getData(db, sql)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getRowJSON(db *sql.DB, sql string) (string, error) {
	results, err := getData(db, sql)
	if err != nil {
		return "", err
	}
	if len(results) > 1 {
		return "", errors.New("query selected more then one row")
	}
	if len(results) < 1 {
		return "", errors.New("query did not select a row")
	}
	data, err := json.Marshal(results[0])
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func setData(db *sql.DB, sql string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(sql)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func connectDB(server, database, username, password string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=require",
		username,
		password,
		server,
		database,
	)
	db, err := sql.Open("postgres", connStr)
	return db, err
}

func toSelect(database, schema, table string, columns []string, where map[string]string, close bool) string {
	var stmt bytes.Buffer
	stmt.WriteString("SELECT")

	// columns
	if len(columns) < 1 {
		stmt.WriteString(" *")
	} else {
		for i, v := range columns {
			if i > 0 {
				stmt.WriteString(fmt.Sprintf(`, "%s"`, v))
			} else {
				stmt.WriteString(fmt.Sprintf(` "%s"`, v))
			}
		}
	}

	// from
	stmt.WriteString(fmt.Sprintf(` FROM "%s"."%s"."%s"`, database, schema, table))

	if len(where) > 0 {
		stmt.WriteString(" WHERE")
		var i uint
		for k, v := range where {
			if i > 0 {
				stmt.WriteString(fmt.Sprintf(`AND "%s" = '%s'`, k, v))
			} else {
				stmt.WriteString(fmt.Sprintf(`"%s" = '%s'`, k, v))
			}
			i++
		}
	}

	if close {
		stmt.WriteString(";")
	}

	return stmt.String()
}

//
// INSERT INTO distributors (did, dname)
// VALUES (5, 'Gizmo Transglobal'), (6, 'Associated Computing, Inc')
// ON CONFLICT (did) DO UPDATE SET dname = EXCLUDED.dname;
//
func toUpsert(database, schema, table string, values, where map[string]string) string {
	var stmt bytes.Buffer

	stmt.WriteString(fmt.Sprintf(`INSERT INTO "%s"."%s"."%s" (`, database, schema, table))

	inserts := make(map[string]string, 0)
	for k, v := range values {
		inserts[k] = v
	}
	for k, v := range where {
		inserts[k] = v
	}

	var i int
	var vals bytes.Buffer
	for k, v := range inserts {
		if i > 0 {
			stmt.WriteString(fmt.Sprintf(`, "%s"`, k))
			vals.WriteString(fmt.Sprintf(`, '%s'`, v))
		} else {
			stmt.WriteString(fmt.Sprintf(`"%s"`, k))
			vals.WriteString(fmt.Sprintf(`'%s'`, v))
		}
		i++
	}
	stmt.WriteString(`) VALUES (` + vals.String() + `) ON CONFLICT (`)

	i = 0
	for k := range where {
		if i > 0 {
			stmt.WriteString(fmt.Sprintf(`, "%s"`, k))
		} else {
			stmt.WriteString(fmt.Sprintf(`"%s"`, k))
		}
		i++
	}
	stmt.WriteString(`) DO UPDATE SET `)

	i = 0
	for k, v := range values {
		if i > 0 {
			stmt.WriteString(fmt.Sprintf(`, "%s" = '%s'`, k, v))
		} else {
			stmt.WriteString(fmt.Sprintf(`"%s" = '%s'`, k, v))
		}
		i++
	}
	stmt.WriteString(";")

	sql := stmt.String()
	fmt.Println(sql) //remove
	return sql
}

func parseJSONRowRequest(r *http.Request) (map[string]string, error) {
	body, ok := getRequestBody(r)
	if !ok {
		return nil, errors.New("could not parse request body")
	}
	if len(body) < 1 {
		return nil, errors.New("request body was empty")
	}

	row := make(map[string]string, 0)
	err := json.Unmarshal(body, &row)
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("JSON object was empty")
	}

	return row, nil
}
