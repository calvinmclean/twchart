package thermoworksbread

import (
	"fmt"
	"log"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	prefermentColor  = "rgba(144, 238, 144, 0.4)"
	bulkFermentColor = "rgba(255, 255, 102, 0.4)"
	finalProofColor  = "rgba(173, 216, 230, 0.4)"
)

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

	AmbientProbePosition ProbePosition
	OvenProbePosition    ProbePosition
	DoughProbePosition   ProbePosition
	OtherProbePosition   ProbePosition
}

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

func (bd BreadData) ChartData() (
	ambientTemperature []opts.LineData,
	ovenTemperature []opts.LineData,
	doughTemperature []opts.LineData,
	other []opts.LineData,
) {
	for _, data := range bd.Data {
		ambientTemperature = data.appendProbeData(ambientTemperature, bd.AmbientProbePosition)
		ovenTemperature = data.appendProbeData(ovenTemperature, bd.OvenProbePosition)
		doughTemperature = data.appendProbeData(doughTemperature, bd.DoughProbePosition)
		other = data.appendProbeData(other, bd.OtherProbePosition)
	}

	return
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
