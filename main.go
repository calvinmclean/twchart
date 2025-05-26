package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type ThermoworksData struct {
	Time      time.Time
	ProbeData []float64
}

func iterCSV(filename string) (iter.Seq2[ThermoworksData, error], func() error, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}

	if len(headers) < 2 || headers[0] != "DateTime" {
		_ = file.Close()
		return nil, nil, fmt.Errorf("unexpected header format")
	}

	return func(yield func(ThermoworksData, error) bool) {
		for {
			record, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				if !yield(ThermoworksData{}, err) {
					return
				}
				continue
			}

			if len(record) < 1 {
				continue
			}

			dt, err := time.ParseInLocation(time.DateTime, record[0], time.Local)
			if err != nil {
				if !yield(ThermoworksData{}, err) {
					return
				}
				continue
			}
			dt = dt.Local()

			probes := make([]float64, len(headers)-1)
			for i := 1; i < len(headers); i++ {
				if record[i] == "" {
					probes[i-1] = -1 // or math.NaN() if you prefer
					continue
				}

				val, err := strconv.ParseFloat(record[i], 64)
				if err != nil {
					if !yield(ThermoworksData{}, err) {
						return
					}
					continue
				}
				probes[i-1] = val
			}

			data := ThermoworksData{
				Time: dt, ProbeData: probes,
			}
			if !yield(data, nil) {
				return
			}
		}
	}, file.Close, nil
}

type Stage struct {
	Name     string
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

func (s *Stage) Finish(t time.Time) {
	s.End = t
	s.Duration = s.End.Sub(s.Start)
}

func (s *Stage) MarkArea(color string) []opts.MarkAreaData {
	return []opts.MarkAreaData{
		{
			Name:  fmt.Sprintf("%s (%s)", s.Name, s.Duration),
			XAxis: s.Start,
			MarkAreaStyle: opts.MarkAreaStyle{
				ItemStyle: &opts.ItemStyle{
					Color: color,
				},
				Label: &opts.Label{
					Show: opts.Bool(true),
				},
			},
		},
		{
			XAxis: s.End,
		},
	}
}

type Event struct {
	Name string
	Time time.Time
}

type BreadData struct {
	Name        string
	Preferment  Stage
	BulkFerment Stage
	FinalProof  Stage

	Events []Event

	Data []ThermoworksData
}

const (
	prefermentColor  = "rgba(144, 238, 144, 0.4)"
	bulkFermentColor = "rgba(255, 255, 102, 0.4)"
	finalProofColor  = "rgba(173, 216, 230, 0.4)"
)

func (bd *BreadData) StartPreferment(t time.Time, name string) {
	if name == "" {
		name = "Preferment"
	}
	bd.Preferment = Stage{
		Name:  name,
		Start: t,
	}
}

func (bd *BreadData) StartBulkFerment(t time.Time, name string) {
	if name == "" {
		name = "Bulk Fermentation"
	}
	bd.BulkFerment = Stage{
		Name:  name,
		Start: t,
	}
	bd.Preferment.Finish(t)
}

func (bd *BreadData) StartFinalProof(t time.Time, name string) {
	if name == "" {
		name = "Final Proof"
	}
	bd.FinalProof = Stage{
		Name:  name,
		Start: t,
	}
	bd.BulkFerment.Finish(t)
}

func (bd *BreadData) EndFinalProof(t time.Time) {
	bd.FinalProof.Finish(t)
}

func (bd *BreadData) AddEvents(event ...Event) {
	bd.Events = append(bd.Events, event...)
}

// Fill in the gaps of time if some stages don't have start/end times
func (bd *BreadData) SetEmptyTimes() {
	// Start by setting end on Preferment
	if bd.Preferment.End.IsZero() && !bd.Preferment.Start.IsZero() && bd.Preferment.Duration > 0 {
		bd.Preferment.End = bd.Preferment.Start.Add(bd.Preferment.Duration)
	}
	if bd.Preferment.End.IsZero() && !bd.BulkFerment.Start.IsZero() {
		bd.Preferment.End = bd.BulkFerment.Start
	}

	// Then set start on BulkFerment
	if bd.BulkFerment.Start.IsZero() && !bd.Preferment.End.IsZero() {
		bd.BulkFerment.Start = bd.Preferment.End
	}
	// and end on BulkFerment
	if bd.BulkFerment.End.IsZero() && !bd.BulkFerment.Start.IsZero() && bd.BulkFerment.Duration > 0 {
		bd.BulkFerment.End = bd.BulkFerment.Start.Add(bd.BulkFerment.Duration)
	}
	if bd.BulkFerment.End.IsZero() && !bd.FinalProof.Start.IsZero() {
		bd.BulkFerment.End = bd.FinalProof.Start
	}

	// Then start of FinalProof
	if bd.FinalProof.Start.IsZero() && !bd.BulkFerment.End.IsZero() {
		bd.FinalProof.Start = bd.BulkFerment.End
	}
	// and end of FinalProof
	if bd.FinalProof.End.IsZero() && !bd.FinalProof.Start.IsZero() && bd.FinalProof.Duration > 0 {
		bd.FinalProof.End = bd.FinalProof.Start.Add(bd.FinalProof.Duration)
	}

	// set durations
	if bd.Preferment.Duration == 0 {
		bd.Preferment.Duration = bd.Preferment.End.Sub(bd.Preferment.Start)
	}
	if bd.BulkFerment.Duration == 0 {
		bd.BulkFerment.Duration = bd.BulkFerment.End.Sub(bd.BulkFerment.Start)
	}
	if bd.FinalProof.Duration == 0 {
		bd.FinalProof.Duration = bd.FinalProof.End.Sub(bd.FinalProof.Start)
	}
}

func (bd *BreadData) LoadData(csvFile string) error {
	csvData, close, err := iterCSV(csvFile)
	if err != nil {
		return err
	}

	for data, err := range csvData {
		if err != nil {
			log.Println("CSV ERR:", err)
			continue
		}

		bd.Data = append(bd.Data, data)
	}

	return close()
}

func (bd BreadData) ChartData() ([]opts.LineData, []opts.LineData, []opts.LineData, []opts.LineData) {
	var probe0Data, probe1Data, probe2Data, probe3Data []opts.LineData

	for _, data := range bd.Data {
		if data.ProbeData[0] > 0 {
			probe0Data = append(probe0Data, opts.LineData{
				Value: []any{data.Time.Format(time.RFC3339), data.ProbeData[0]},
			})
		}
		if data.ProbeData[1] > 0 {
			probe1Data = append(probe1Data, opts.LineData{
				Value: []any{data.Time.Format(time.RFC3339), data.ProbeData[1]},
			})
		}
		if data.ProbeData[2] > 0 {
			probe2Data = append(probe2Data, opts.LineData{
				Value: []any{data.Time.Format(time.RFC3339), data.ProbeData[2]},
			})
		}
		if data.ProbeData[3] > 0 {
			probe3Data = append(probe3Data, opts.LineData{
				Value: []any{data.Time.Format(time.RFC3339), data.ProbeData[3]},
			})
		}
	}

	return probe0Data, probe1Data, probe2Data, probe3Data
}

func (bd BreadData) Chart() (*charts.Line, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Bread Baking Temperatures",
			Subtitle: "tracking temperatures throughout the bread-making process",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "time",
			AxisPointer: &opts.AxisPointer{
				Show: opts.Bool(true),
				Snap: opts.Bool(false),
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "item",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   25,
		}),
	)

	probe0Data, probe1Data, _, _ := bd.ChartData()

	events := []opts.MarkLineNameXAxisItem{}
	for _, event := range bd.Events {
		events = append(events, opts.MarkLineNameXAxisItem{
			Name:  event.Name,
			XAxis: event.Time,
		})
	}

	options := []charts.SeriesOpts{
		charts.WithLineChartOpts(
			opts.LineChart{
				Smooth:     opts.Bool(true),
				ShowSymbol: opts.Bool(false),
			},
		),
		charts.WithMarkLineNameXAxisItemOpts(events...),
		charts.WithMarkLineStyleOpts(opts.MarkLineStyle{
			Symbol: []string{"none", "none"},
			Label: &opts.Label{
				Show:      opts.Bool(true),
				Formatter: " ", // empty
				// Formatter: "{b}",
			},
		}),
		charts.WithMarkAreaData(bd.Preferment.MarkArea(prefermentColor)),
		charts.WithMarkAreaData(bd.BulkFerment.MarkArea(bulkFermentColor)),
		charts.WithMarkAreaData(bd.FinalProof.MarkArea(finalProofColor)),
	}

	line.AddSeries("Ambient Temperature", probe0Data, options...).
		AddSeries("Oven Temperature", probe1Data,
			charts.WithLineChartOpts(
				opts.LineChart{
					Smooth:     opts.Bool(true),
					ShowSymbol: opts.Bool(false),
				},
			),
		)

	return line, nil
}

func createBreadDataRealtime(start time.Time) BreadData {
	bd := BreadData{Name: "Ciabatta"}

	now := start

	bd.AddEvents(Event{Name: "Mix biga", Time: now})
	now = now.Add(3 * time.Minute)

	bd.StartPreferment(now, "Biga Fermentation")
	now = now.Add(11 * time.Hour)

	bd.StartBulkFerment(now, "")
	now = now.Add(1 * time.Hour)

	bd.AddEvents(Event{Name: "12 stretch and folds", Time: now})
	now = now.Add(1 * time.Hour)

	bd.AddEvents(Event{Name: "Shape", Time: now})
	now = now.Add(2 * time.Minute)
	bd.StartFinalProof(now, "")

	now = now.Add(90 * time.Minute)
	bd.EndFinalProof(now)

	bd.AddEvents(Event{Name: "Bake", Time: now})
	now = now.Add(25 * time.Minute)

	bd.AddEvents(Event{Name: "Done", Time: now})

	return bd
}

func main() {
	start := time.Date(2025, time.May, 24, 20, 10, 0, 0, time.Local)
	data := createBreadDataRealtime(start)

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
