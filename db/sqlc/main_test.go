// In order to write the test, We have to setup the Connection and the Queries object first (to do that inside this file)
package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:12345678@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries // Will define a testQueries object(contains a DBTX which is db conn or Tx) as a global variable

/*
 * The main entry point of all unit tests inside 1 specific Golang package (package db)
 */

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource) // Create a new connection to the DB, returns a connection obj and an Error
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(conn) // Use a connection to create a new testQueries object

	// m.Run(): To start running the Unit test which will return an 'exit code'(test pass or fail)
	// Then we should report it back to the test runner via os.Exit() command.
	os.Exit(m.Run())
}
