package database

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func New(config *Config) (*DB, error) {
	db, err := sqlx.Connect("mysql", DSN(config))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db")
	}
	return &DB{db}, nil
}
