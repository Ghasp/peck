package main

import (
	"database/sql"
	"flag"
	"os"
)

// TOKEN global variable
// commandline paramter: --Token
// enviroment variable: PECK_TOKEN
var TOKEN string = ""

// SERVER global variable
// commandline paramter: --Server
// enviroment variable: PECK_SERVER
var SERVER string = "localhost"

// DATABASE global variable
// commandline paramter: --Database
// enviroment variable: PECK_DATABASE
var DATABASE string = "postgres"

// USERNAME global variable
// commandline paramter: --Username
// enviroment variable: PECK_USERNAME
var USERNAME string = ""

// PASSWORD global variable
// commandline paramter: --Password
// enviroment variable: PECK_PASSWORD
var PASSWORD string = ""

// PRIMARYKEY global variable
// commandline paramter: --PrimaryKey
// enviroment variable: PECK_PRIMARY_KEY
var PRIMARYKEY string = ""

// TLSKEYPATH global variable
// commandline paramter: --TlsKeyPath
// enviroment variable: PECK_TLS_KEY_PATH
var TLSKEYPATH string

// TLSCERTPATH global variable
// commandline paramter: --TlsCertPath
// enviroment variable: PECK_TLS_CERT_PATH
var TLSCERTPATH string

// CONNECTION global variable
var CONNECTION *sql.DB

// readCommandlineParameters configures global settings from the specified commandline parameters.
// commandline parameters overrided enviorment variables if they are set.
func readCommandlineParameters() {
	flag.StringVar(&TOKEN, "Token", TOKEN, "The Token to be used by clients to authenticate.")
	flag.StringVar(&SERVER, "Server", SERVER, "The Postgres server host name.")
	flag.StringVar(&DATABASE, "Database", DATABASE, "The Postgres database name.")
	flag.StringVar(&USERNAME, "Username", USERNAME, "The Postgres username that will be used to perform sql queries.")
	flag.StringVar(&PASSWORD, "Password", PASSWORD, "The Postgres password the specified username.")
	flag.StringVar(&PRIMARYKEY, "PrimaryKey", PRIMARYKEY, "The column name of all table primary key which will be used by the '/schema/table/key' interface. if left blank the interface will be disabled. this interface only makes sense to use when the database has a naming convention for primary key columns such as 'id'.")
	flag.StringVar(&TLSKEYPATH, "TlsKeyPath", TLSKEYPATH, "The TLS Key path to enabled https")
	flag.StringVar(&TLSCERTPATH, "TlsCertPath", TLSCERTPATH, "The TLS Cert path to enabled https")
	flag.Parse()
}

// readEnviormentVarialbes configures global settings from the specified enviorment variable.
// All enviorment variables can be overrided by command line arguments.
func readEnviormentVarialbes() {
	TOKEN = os.Getenv("PECK_TOKEN")
	SERVER = os.Getenv("PECK_SERVER")
	DATABASE = os.Getenv("PECK_DATABASE")
	USERNAME = os.Getenv("PECK_USERNAME")
	PASSWORD = os.Getenv("PECK_PASSWORD")
	PRIMARYKEY = os.Getenv("PECK_PRIMARY_KEY")
	PRIMARYKEY = os.Getenv("PECK_TLS_KEY_PATH")
	PRIMARYKEY = os.Getenv("PECK_TLS_CERT_PATH")
}
