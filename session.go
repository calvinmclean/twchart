package thermoworksbread

import (
	"fmt"
	"log"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

type Session struct {
	Name string
	Date time.Time

	Probes []Probe
	Stages []Stage
	Events []Event

	Data []ThermoworksData
}

type Stage struct {
	Name     string
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

type Event struct {
	Note string
	Time time.Time
}

type Probe struct {
	Name     string
	Position ProbePosition
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

func (s *Session) LoadData(csvFile string) error {
	csvData, close, err := iterCSV(csvFile)
	if err != nil {
		return err
	}

	for data, err := range csvData {
		if err != nil {
			log.Println("CSV ERR:", err)
			continue
		}

		s.Data = append(s.Data, data)
	}

	return close()
}
