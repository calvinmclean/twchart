package main

import (
	"io"
	"log"
	"net/http"
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

	bd.AddEvents(thermoworksbread.Event{Name: "Mix biga", Time: now})
	now = now.Add(3 * time.Minute)

	bd.StartPreferment(now, "Biga Fermentation")
	now = now.Add(11 * time.Hour)

	bd.StartBulkFerment(now, "")
	now = now.Add(1 * time.Hour)

	bd.AddEvents(thermoworksbread.Event{Name: "12 stretch and folds", Time: now})
	now = now.Add(1 * time.Hour)

	bd.AddEvents(thermoworksbread.Event{Name: "Shape", Time: now})
	now = now.Add(2 * time.Minute)
	bd.StartFinalProof(now, "")

	now = now.Add(90 * time.Minute)
	bd.EndFinalProof(now)

	bd.AddEvents(thermoworksbread.Event{Name: "Bake", Time: now})
	now = now.Add(25 * time.Minute)

	bd.AddEvents(thermoworksbread.Event{Name: "Done", Time: now})

	return bd
}

func main() {
	start := time.Date(2025, time.May, 24, 20, 10, 0, 0, time.Local)
	data := createExampleBreadData(start)

	log.Println("running server at http://localhost:8089")
	log.Fatal(http.ListenAndServe("localhost:8089", logRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := data.LoadData("chart.csv")
		if err != nil {
			panic(err)
		}

		chart, err := data.Chart()
		if err != nil {
			panic(err)
		}

		page := components.NewPage()
		page.AddCharts(chart)
		page.Render(io.MultiWriter(w))
	}))))
}
