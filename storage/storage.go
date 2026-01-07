package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/calvinmclean/twchart/storage/db"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate sqlc generate

//go:embed schema.sql
var ddl string

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

	ddlQueries := strings.SplitSeq(ddl, "\n\n")
	for q := range ddlQueries {
		_, err = database.Exec(q)
		if err != nil &&
			!strings.Contains(err.Error(), "duplicate column name:") && // allow "idempotent" migration
			!strings.Contains(err.Error(), "no such column:") {
			return nil, fmt.Errorf("error creating tables: %w", err)
		}
	}

	return &Client{
		db.New(database),
		database,
	}, nil
}

func (c Client) Close() {
	c.db.Close()
}
