package thermoworksbread

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
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
		Name:  "Bulk Fermentation",
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
			Name:  event.Note,
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
		charts.WithMarkAreaData(bd.Bake.MarkArea(bakeColor)),
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

func ParseBreadData(input []byte) (BreadData, error) {
	bd := BreadData{}

	var currentDate time.Time
	scanner := bufio.NewScanner(bytes.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if bd.Name == "" {
			bd.Name = line
			continue
		}

		parts := strings.SplitN(line, ":", 2)

		if len(parts) != 2 {
			return bd, fmt.Errorf("unexpected format: %v", parts)
		}

		value := strings.TrimSpace(parts[1])

		var err error
		switch strings.ToLower(parts[0]) {
		case "date":
			currentDate, err = time.ParseInLocation(time.DateOnly, value, time.Local)
			if err != nil {
				return bd, fmt.Errorf("error parsing date: %w", err)
			}
		case "note":
			parsedTime, note, err := parseNote(value, currentDate)
			if err != nil {
				return bd, fmt.Errorf("error parsing note: %w", err)
			}
			bd.AddEvents(Event{Time: parsedTime, Note: note})
			currentDate = parsedTime
		case "preferment":
			currentDate, err = parseStage(currentDate, value, bd.StartPreferment)
			if err != nil {
				return bd, fmt.Errorf("error parsing preferment: %w", err)
			}
		case "bulk ferment":
			currentDate, err = parseStage(currentDate, value, bd.StartBulkFerment)
			if err != nil {
				return bd, fmt.Errorf("error parsing bulk ferment: %w", err)
			}
		case "final proof":
			currentDate, err = parseStage(currentDate, value, bd.StartFinalProof)
			if err != nil {
				return bd, fmt.Errorf("error parsing final proof: %w", err)
			}
		case "bake":
			currentDate, err = parseStage(currentDate, value, bd.StartBake)
			if err != nil {
				return bd, fmt.Errorf("error parsing bake: %w", err)
			}
		case "done":
			currentDate, err = parseStage(currentDate, value, bd.EndBake)
			if err != nil {
				return bd, fmt.Errorf("error parsing bake: %w", err)
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		return bd, fmt.Errorf("error scanning: %w", err)
	}

	return bd, nil
}

func parseNote(value string, currentTime time.Time) (time.Time, string, error) {
	// Find the second colon in " 6:53PM: ..."
	second := false
	i := strings.IndexFunc(value, func(r rune) bool {
		if r != ':' {
			return false
		}
		if !second {
			second = true
			return false
		}
		return true
	})
	if i < 0 {
		fmt.Println(value)
		return time.Time{}, "", errors.New("note is missing expected number of ':'")
	}

	parsedTime, err := parseTime(strings.TrimSpace(value[0:i]), currentTime)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing time: %w", err)
	}

	if isNextDay(currentTime, parsedTime) {
		parsedTime = parsedTime.AddDate(0, 0, 1)
	}

	note := strings.TrimSpace(value[i+1:])
	return parsedTime, note, nil
}

func isNextDay(currentTime, newTime time.Time) bool {
	// if currentTime is PM and newTime is AM, it is the next day
	return currentTime.Hour() >= 12 && newTime.Hour() < 12
}

func parseStage(currentTime time.Time, value string, handle func(time.Time)) (time.Time, error) {
	parsedTime, err := parseTime(value, currentTime)
	if err != nil {
		return time.Time{}, err
	}
	if isNextDay(currentTime, parsedTime) {
		parsedTime = parsedTime.AddDate(0, 0, 1)
	}
	handle(parsedTime)
	return parsedTime, nil
}

func parseTime(input string, date time.Time) (time.Time, error) {
	result, err := time.ParseInLocation(time.Kitchen, input, time.Local)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		result.Hour(),
		result.Minute(),
		0, 0, time.Local,
	), nil
}
