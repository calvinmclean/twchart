package twchart

import (
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func (s Session) ChartData() [][]opts.LineData {
	result := make([][]opts.LineData, len(s.Probes))

	if len(s.Probes) == 0 {
		return result
	}

	for _, datum := range s.Data {
		for _, p := range s.Probes {
			probeData := datum.GetProbeData(p.Position)
			if probeData <= 0 {
				result[p.Position-1] = append(result[p.Position-1], opts.LineData{
					Value: []any{datum.Time.Format(time.RFC3339), nil},
				})
			}

			result[p.Position-1] = append(result[p.Position-1], opts.LineData{
				Value: []any{datum.Time.Format(time.RFC3339), probeData},
			})
		}
	}

	// Add time bounds so all Events and Stages show
	earliest, latest := s.TimeBounds()
	result[0] = slices.Insert(result[0], 0, opts.LineData{
		Value: []any{earliest.Format(time.RFC3339), nil},
	})
	result[0] = append(result[0], opts.LineData{
		Value: []any{latest.Format(time.RFC3339), nil},
	})

	return result
}

func (s Session) Chart() (*charts.Line, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "90%",
			Height: "80vh",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Temperatures",
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
		charts.WithDataZoomOpts(
			opts.DataZoom{
				Type:   "slider",
				Start:  0,
				End:    25,
				Orient: "horizontal",
			},
			opts.DataZoom{
				Type:   "slider",
				Start:  0,
				End:    100,
				Orient: "vertical",
			},
		),
	)

	events := []opts.MarkLineNameXAxisItem{}
	for _, event := range s.Events {
		events = append(events, opts.MarkLineNameXAxisItem{
			Name:  event.Note,
			XAxis: event.Time,
		})
	}

	baseOpts := []charts.SeriesOpts{
		charts.WithLineChartOpts(
			opts.LineChart{
				Smooth:       opts.Bool(true),
				ShowSymbol:   opts.Bool(false),
				ConnectNulls: opts.Bool(false),
			},
		),
	}

	colors := []string{
		"rgba(144, 238, 144, 0.4)",
		"rgba(255, 255, 102, 0.4)",
		"rgba(173, 216, 230, 0.4)",
		"rgba(255, 173, 177, 0.4)",
	}
	areas := []charts.SeriesOpts{}
	for i, stage := range s.Stages {
		areas = append(areas, charts.WithMarkAreaData(stage.MarkArea(colors[i])))
	}

	optsWithAreaAndEvents := append(baseOpts,
		charts.WithMarkLineNameXAxisItemOpts(events...),
		charts.WithMarkLineStyleOpts(opts.MarkLineStyle{
			Symbol: []string{"none", "none"},
			Label: &opts.Label{
				Show:      opts.Bool(true),
				Formatter: " ", // empty
				// Formatter: "{b}",
			},
		}),
	)
	optsWithAreaAndEvents = append(optsWithAreaAndEvents, areas...)

	chartData := s.ChartData()
	for i, probe := range s.Probes {
		opts := baseOpts
		if i == 0 {
			opts = optsWithAreaAndEvents
		}
		line.AddSeries(probe.Name, chartData[probe.Position-1], opts...)
	}

	return line, nil
}
