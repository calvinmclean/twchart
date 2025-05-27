package thermoworksbread

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

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
