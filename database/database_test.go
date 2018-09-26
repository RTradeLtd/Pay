package database_test

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
)

var (
	travis = os.Getenv("TRAVIS") != ""
	dbPass string
)

func TestDatabase(t *testing.T) {
	db, err := gorm.Open(
		"postgres", "host=127.0.0.1 port=5433 user=postgres dbname=temporal password=password123 sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}
