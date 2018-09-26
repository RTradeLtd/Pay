package database

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DatabaseManager struct {
	DB *gorm.DB
}

func Initialize(cfg *config.TemporalConfig) (*DatabaseManager, error) {
	if cfg == nil {
		return nil, errors.New("invalid configuration provided")
	}

	db, err := OpenDBConnection(DBOptions{
		User:     cfg.Database.Name,
		Password: cfg.Database.Password,
		Address:  cfg.Database.URL,
	})
	if err != nil {
		return nil, err
	}

	dbm := DatabaseManager{DB: db}
	return &dbm, nil
}

// DBOptions declares options for opening a database connection
type DBOptions struct {
	User           string
	Password       string
	Address        string
	SSLModeDisable bool
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection(opts DBOptions) (*gorm.DB, error) {
	if opts.User == "" {
		opts.User = "postgres"
	}
	// look into whether or not we wil disable sslmode
	dbConnURL := fmt.Sprintf("host=%s port=5433 user=%s dbname=temporal password=%s",
		opts.Address, opts.User, opts.Password)
	if opts.SSLModeDisable {
		dbConnURL = "sslmode=disable " + dbConnURL
	}
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func OpenTestDBConnection(dbPass string) (*gorm.DB, error) {
	dbConnURL := fmt.Sprintf("host=127.0.0.1 port=5433 user=postgres dbname=temporal password=%s", dbPass)
	db, err := gorm.Open("postgres", dbConnURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) error {
	err := db.Close()
	if err != nil {
		return err
	}
	return nil
}
