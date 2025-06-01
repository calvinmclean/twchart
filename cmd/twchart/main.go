package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/calvinmclean/twchart"
	"github.com/go-echarts/go-echarts/v2/charts"
)

func main() {
	var filename, dir string
	flag.StringVar(&filename, "file", "", "filename to read BreadData from")
	flag.StringVar(&dir, "dir", "", "directory to read BreadData from")
	flag.Parse()

	var h http.Handler
	switch {
	case dir != "":
		err := writeChartsInDir(dir)
		if err != nil {
			log.Fatalf("failed to create charts for directory: %v", err)
		}

		h = http.FileServer(http.Dir(dir))
	case filename != "":
		chart, err := createChartForFile(filename)
		if err != nil {
			log.Fatalf("error creating chart for file %q: %v", filename, err)
		}

		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := chart.Render(w)
			if err != nil {
				log.Fatalf("error rendering chart: %v", err)
			}
		})
	}

	log.Println("running server at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(h)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func writeChartsInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		filename := filepath.Join(dir, entry.Name())

		chart, err := createChartForFile(filename)
		if err != nil {
			return fmt.Errorf("error creating chart for %q: %v", filename, err)
		}

		htmlFileName := strings.TrimSuffix(entry.Name(), ".txt") + ".html"
		htmlPath := filepath.Join(dir, htmlFileName)
		f, err := os.Create(htmlPath)
		if err != nil {
			return fmt.Errorf("error creating HTML file %q: %v", htmlPath, err)
		}
		defer f.Close()

		err = chart.Render(f)
		if err != nil {
			return fmt.Errorf("error rendering chart to file %q: %v", htmlPath, err)
		}
	}

	return nil
}

func createChartForFile(filename string) (*charts.Line, error) {
	var s twchart.Session

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", filename, err)
	}
	defer f.Close()

	_, err = io.Copy(&s, f)
	if err != nil {
		return nil, fmt.Errorf("error parsing BreadData: %w", err)
	}

	dataFilename := strings.TrimSuffix(filename, ".txt") + ".csv"
	err = s.LoadData(dataFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading Thermoworks data: %w", err)
	}

	chart, err := s.Chart()
	if err != nil {
		return nil, fmt.Errorf("error creating chart: %w", err)
	}

	return chart, nil
}
