package thermoworksbread

import (
	"fmt"
	"log"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	prefermentColor  = "rgba(144, 238, 144, 0.4)"
	bulkFermentColor = "rgba(255, 255, 102, 0.4)"
	finalProofColor  = "rgba(173, 216, 230, 0.4)"
	bakeColor        = "rgba(255, 173, 177, 0.4)"
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
	Note string
	Time time.Time
}

type BreadData struct {
	Name        string
	Date        time.Time
	Preferment  Stage
	BulkFerment Stage
	FinalProof  Stage
	Bake        Stage

	Events []Event

	Data []ThermoworksData

	AmbientProbePosition ProbePosition
	OvenProbePosition    ProbePosition
	DoughProbePosition   ProbePosition
	OtherProbePosition   ProbePosition
}

func (bd *BreadData) StartPreferment(t time.Time) {
	bd.Preferment = Stage{
		Name:  "Preferment",
		Start: t,
	}
}

func (bd *BreadData) StartBulkFerment(t time.Time) {
	bd.BulkFerment = Stage{
		Name:  "Bulk Ferment",
		Start: t,
	}
	bd.Preferment.Finish(t)
}

func (bd *BreadData) StartFinalProof(t time.Time) {
	bd.FinalProof = Stage{
		Name:  "Final Proof",
		Start: t,
	}
	bd.BulkFerment.Finish(t)
}

func (bd *BreadData) StartBake(t time.Time) {
	bd.Bake = Stage{
		Name:  "Bake",
		Start: t,
	}
	bd.FinalProof.Finish(t)
}

func (bd *BreadData) EndBake(t time.Time) {
	bd.Bake.Finish(t)
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
	if bd.FinalProof.End.IsZero() && !bd.Bake.Start.IsZero() {
		bd.FinalProof.End = bd.Bake.Start
	}

	// Then start of Bake
	if bd.Bake.Start.IsZero() && !bd.FinalProof.End.IsZero() {
		bd.Bake.Start = bd.FinalProof.End
	}
	// and end of Bake
	if bd.Bake.End.IsZero() && !bd.Bake.Start.IsZero() && bd.Bake.Duration > 0 {
		bd.Bake.End = bd.Bake.Start.Add(bd.Bake.Duration)
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
	if bd.Bake.Duration == 0 {
		bd.Bake.Duration = bd.Bake.End.Sub(bd.Bake.Start)
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
