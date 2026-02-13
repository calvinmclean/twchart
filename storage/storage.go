package storage

import (
	"database/sql"
	"fmt"

	"github.com/calvinmclean/twchart/storage/db"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate sqlc generate

type Client struct {
	*db.Queries
	db *sql.DB
}

func New(filename string) (*Client, error) {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &Client{
		db.New(database),
		database,
	}, nil
}

func (c Client) Close() {
	c.db.Close()
}
