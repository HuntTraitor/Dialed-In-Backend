package data

import (
	"database/sql"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {

	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal(err)
	}

	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	db, err := sql.Open("postgres", testDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("../../db/sql/test_setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("../../db/sql/test_teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

	})
	return db
}
