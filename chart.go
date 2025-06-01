package thermoworksbread

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func (bd BreadData) ChartData() [][]opts.LineData {
	result := make([][]opts.LineData, len(bd.Probes))
	for _, datum := range bd.Data {
		for _, p := range bd.Probes {
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

	return result
}

func (bd BreadData) Chart() (*charts.Line, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "90%",
			Height: "80vh",
		}),
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

	events := []opts.MarkLineNameXAxisItem{}
	for _, event := range bd.Events {
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
	for i, stage := range bd.Stages {
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

	chartData := bd.ChartData()
	for i, probe := range bd.Probes {
		opts := baseOpts
		if i == 0 {
			opts = optsWithAreaAndEvents
		}
		line.AddSeries(probe.Name, chartData[probe.Position-1], opts...)
	}

	return line, nil
}
