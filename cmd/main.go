package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	thermoworksbread "github.com/calvinmclean/thermoworks-bread"
	"github.com/go-echarts/go-echarts/v2/components"
)

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func createExampleBreadData(start time.Time) thermoworksbread.BreadData {
	bd := thermoworksbread.BreadData{
		Name:                 "Ciabatta",
		AmbientProbePosition: thermoworksbread.ProbePosition1,
		OvenProbePosition:    thermoworksbread.ProbePosition2,
	}

	now := start

	bd.AddEvents(thermoworksbread.Event{Note: "Mix biga", Time: now})
	now = now.Add(3 * time.Minute)

	bd.StartPreferment(now)
	now = now.Add(11 * time.Hour)

	bd.StartBulkFerment(now)
	now = now.Add(1 * time.Hour)

	bd.AddEvents(thermoworksbread.Event{Note: "12 stretch and folds", Time: now})
	now = now.Add(1 * time.Hour)

	bd.AddEvents(thermoworksbread.Event{Note: "Shape", Time: now})
	now = now.Add(2 * time.Minute)
	bd.StartFinalProof(now)
	now = now.Add(90 * time.Minute)

	bd.StartBake(now)
	now = now.Add(25 * time.Minute)
	bd.EndBake(now)

	bd.AddEvents(thermoworksbread.Event{Note: "Done", Time: now})

	return bd
}

func main() {
	var filename string
	var example, stdin bool
	flag.StringVar(&filename, "file", "", "filename to read BreadData from")
	flag.BoolVar(&example, "example", false, "use example data")
	flag.BoolVar(&stdin, "stdin", false, "read from stdin")
	flag.Parse()

	var bd thermoworksbread.BreadData
	switch {
	case filename != "":
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("error opening file %q: %v", filename, err)
		}
		defer f.Close()

		_, err = io.Copy(&bd, f)
		if err != nil {
			log.Fatalf("error parsing BreadData: %v", err)
		}
	case stdin:
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("error reading stdin: %v", err)
		}

		err = bd.UnmarshalText(input)
		if err != nil {
			log.Fatalf("error parsing BreadData: %v", err)
		}
	case example:
		start := time.Date(2025, time.May, 24, 20, 10, 0, 0, time.Local)
		bd = createExampleBreadData(start)
	}

	err := bd.LoadData("chart.csv")
	if err != nil {
		panic(err)
	}

	log.Println("running server at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		chart, err := bd.Chart()
		if err != nil {
			panic(err)
		}

		page := components.NewPage()
		page.AddCharts(chart)
		page.Render(io.MultiWriter(w))
	}))))
}
