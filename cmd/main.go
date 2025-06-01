package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/calvinmclean/twchart"
	"github.com/go-echarts/go-echarts/v2/components"
)

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	var filename string
	var stdin bool
	flag.StringVar(&filename, "file", "", "filename to read BreadData from")
	flag.BoolVar(&stdin, "stdin", false, "read from stdin")
	flag.Parse()

	var s twchart.Session
	switch {
	case filename != "":
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("error opening file %q: %v", filename, err)
		}
		defer f.Close()

		_, err = io.Copy(&s, f)
		if err != nil {
			log.Fatalf("error parsing BreadData: %v", err)
		}
	case stdin:
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("error reading stdin: %v", err)
		}

		err = s.UnmarshalText(input)
		if err != nil {
			log.Fatalf("error parsing BreadData: %v", err)
		}
	}

	chartFilename := "chart.csv"
	if filename != "" {
		chartFilename = strings.ReplaceAll(filename, ".txt", ".csv")
	}
	err := s.LoadData(chartFilename)
	if err != nil {
		panic(err)
	}

	log.Println("running server at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		chart, err := s.Chart()
		if err != nil {
			panic(err)
		}

		page := components.NewPage()
		page.AddCharts(chart)
		page.Render(io.MultiWriter(w))
	}))))
}
