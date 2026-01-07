package api

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/calvinmclean/twchart"
)

// Load data from files and store in the API's store
func (a *API) Load(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %v", path, err)
		}

		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			return nil
		}

		session, err := loadSessionFromFile(path)
		if err != nil {
			return fmt.Errorf("error creating chart for %q: %v", path, err)
		}

		s := &sessionResource{Session: session}
		fmt.Printf("Loaded %s/%s\n", s.GetID(), s.Session.Name)
		err = a.Storage.Set(context.Background(), s)
		if err != nil {
			return fmt.Errorf("error storing session from %q: %v", path, err)
		}

		return nil
	})
}

func loadSessionFromFile(filename string) (twchart.Session, error) {
	var s twchart.Session

	f, err := os.Open(filename)
	if err != nil {
		return s, fmt.Errorf("error opening file %q: %w", filename, err)
	}
	defer f.Close()

	_, err = io.Copy(&s, f)
	if err != nil {
		return s, fmt.Errorf("error parsing Session: %w", err)
	}

	dataFilename := strings.TrimSuffix(filename, ".txt") + ".csv"
	err = s.LoadDataFromFile(dataFilename)
	if err != nil {
		return s, fmt.Errorf("error loading Thermoworks data: %w", err)
	}

	return s, nil
}
