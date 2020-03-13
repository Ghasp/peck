package main

import "database/sql"

// TOKEN global variable
// commandline paramter: --token
// enviroment variable: PECK_TOKEN
var TOKEN string = ""

// SERVER global variable
// commandline paramter: --server
// enviroment variable: PECK_SERVER
var SERVER string = ""

// DATABASE global variable
// commandline paramter: --database
// enviroment variable: PECK_DATABASE
var DATABASE string = ""

// USERNAME global variable
// commandline paramter: --username
// enviroment variable: PECK_USERNAME
var USERNAME string = ""

// PASSWORD global variable
// commandline paramter: --password
// enviroment variable: PECK_PASSWORD
var PASSWORD string = ""

// PRIMARYKEY global variable
// commandline paramter: --primarykey
// enviroment variable: PECK_PRIMARYKEY
var PRIMARYKEY string = ""

// CONNECTION global variable
var CONNECTION *sql.DB

// readCommandlineParameters configures global settings from the specified commandline parameters.
// commandline parameters overrided enviorment variables if they are set.
func readCommandlineParameters() {

}

// readEnviormentVarialbes configures global settings from the specified enviorment variable.
// All enviorment variables can be overrided by command line arguments.
func readEnviormentVarialbes() {

}
