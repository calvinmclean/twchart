package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/extensions"
	"github.com/calvinmclean/twchart"
	"github.com/calvinmclean/twchart/api"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

func main() {
	server := api.New()
	cmd := server.Command()

	// Enable data loading and storage setup for serve command
	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		if c.Name() != "serve" {
			return nil
		}

		storeFlag := c.Flag("store")
		if storeFlag != nil && storeFlag.Value.String() != "" {
			err := server.Setup(storeFlag.Value.String())
			if err != nil {
				return fmt.Errorf("error setting up storage: %w", err)
			}
		}

		dirFlag := c.Flag("dir")
		if dirFlag == nil || dirFlag.Value.String() == "" {
			return nil
		}

		return server.Load(dirFlag.Value.String())
	}

	// Add custom flags to serve command
	for _, c := range cmd.Commands() {
		if c.Name() != "serve" {
			continue
		}

		c.Flags().String("dir", "", "directory to read data from")
		c.Flags().String("store", "", "filename for JSON KV store")
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate [from] [to]",
		Short: "Migrate from one storage file to another",
		Args:  cobra.ExactArgs(2),
		RunE:  migrateCommand,
	}
	migrateCmd.Flags().Bool("old", false, "use old format when reading the 'from' data")
	cmd.AddCommand(migrateCmd)

	dbMigrateCmd := &cobra.Command{
		Use:   "db-migrate",
		Short: "Run database schema migrations using golang-migrate",
		Long:  "Run database schema migrations up or down using golang-migrate. Defaults to running all up migrations.",
		RunE:  dbMigrateCommand,
	}
	dbMigrateCmd.Flags().String("database", "data.db", "SQLite database file path")
	dbMigrateCmd.Flags().String("migrations", "github://calvinmclean/twchart/migrations#main", "Migration source: local path (e.g., 'migrations') or GitHub URL")
	dbMigrateCmd.Flags().Bool("down", false, "Run down migrations instead of up")
	dbMigrateCmd.Flags().Int("steps", 0, "Number of migrations to run (0 = all)")
	cmd.AddCommand(dbMigrateCmd)

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func migrateCommand(cmd *cobra.Command, args []string) error {
	fromFile := args[0]
	toFile := args[1]

	old, _ := cmd.Flags().GetBool("old")

	ctx := cmd.Context()

	var fromSessions []*api.SessionResource
	if old {
		var err error
		fromSessions, err = parseOld(fromFile)
		if err != nil {
			return fmt.Errorf("error reading old data format: %w", err)
		}
	} else {
		fromAPI := api.New()
		err := fromAPI.Setup(fromFile)
		if err != nil {
			return fmt.Errorf("error setting up JSON storage: %w", err)
		}

		fromSessions, err = fromAPI.Storage.Search(ctx, "", nil)
		if err != nil {
			return fmt.Errorf("error reading sessions from JSON: %w", err)
		}
	}

	toAPI := api.New()
	err := toAPI.Setup(toFile)
	if err != nil {
		return fmt.Errorf("error setting up SQL storage: %w", err)
	}

	for _, session := range fromSessions {
		err := toAPI.Storage.Set(ctx, session)
		if err != nil {
			return fmt.Errorf("error migrating session %s: %w", session.GetID(), err)
		}
		fmt.Printf("Migrated session: %s\n", session.Session.Name)
	}

	fmt.Printf("Successfully migrated %d sessions\n", len(fromSessions))
	return nil
}

// parseOld converts from the old format ({"id": "", "Session": {}, "UploadedAt": ""}) to the new
func parseOld(fname string) ([]*api.SessionResource, error) {
	db, err := extensions.KVConnectionConfig{Filename: fname}.CreateDB()
	if err != nil {
		return nil, fmt.Errorf("error creating hord DB: %w", err)
	}

	keys, err := db.Keys()
	if err != nil {
		return nil, fmt.Errorf("error reading keys: %w", err)
	}

	type oldSession struct {
		ID         string `json:"id"`
		Session    twchart.Session
		UploadedAt time.Time
	}

	var out []*api.SessionResource
	for _, k := range keys {
		b, err := db.Get(k)
		if err != nil {
			return nil, fmt.Errorf("error reading key %q: %w", k, err)
		}

		var s oldSession
		err = json.Unmarshal(b, &s)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling data for key %q: %w", k, err)
		}

		id, err := xid.FromString(s.ID)
		if err != nil {
			return nil, fmt.Errorf("error parsing ID: %w", err)
		}
		s.Session.ID = babyapi.ID{ID: id}
		s.Session.UploadedAt = s.UploadedAt

		out = append(out, &api.SessionResource{
			Session: s.Session,
		})
	}

	return out, nil
}

func dbMigrateCommand(cmd *cobra.Command, _ []string) error {
	dbPath, _ := cmd.Flags().GetString("database")
	migrationsPath, _ := cmd.Flags().GetString("migrations")
	down, _ := cmd.Flags().GetBool("down")
	steps, _ := cmd.Flags().GetInt("steps")

	dbURL := fmt.Sprintf("sqlite3://%s", dbPath)
	migrationsURL := parseMigrationsURL(migrationsPath)

	m, err := migrate.New(migrationsURL, dbURL)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}

	if down {
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running down migrations: %w", err)
		}
		fmt.Println("Down migrations completed successfully")
	} else {
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running up migrations: %w", err)
		}
		fmt.Println("Up migrations completed successfully")
	}

	if err == migrate.ErrNoChange {
		fmt.Println("No migrations to run")
	}

	return nil
}

func parseMigrationsURL(path string) string {
	if path == "" {
		return "github://calvinmclean/twchart/migrations#main"
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return convertGitHubURL(path)
	}

	if strings.HasPrefix(path, "github://") {
		return path
	}

	return fmt.Sprintf("file://%s", path)
}

func convertGitHubURL(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	parts := strings.SplitN(url, "/", 4)
	if len(parts) < 3 {
		return url
	}

	owner := parts[0]
	repo := parts[1]
	ref := "main"
	path := "migrations"

	if len(parts) >= 4 && strings.HasPrefix(parts[2], "tree") {
		pathParts := strings.SplitN(parts[3], "/", 2)
		ref = pathParts[0]
		if len(pathParts) > 1 {
			path = pathParts[1]
		}
	}

	return fmt.Sprintf("github://%s/%s/%s#%s", owner, repo, path, ref)
}
