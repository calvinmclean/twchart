package thermoworksbread

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"os"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

type ProbePosition uint

const (
	ProbePositionNone = iota
	ProbePosition1
	ProbePosition2
	ProbePosition3
	ProbePosition4
)

type ThermoworksData struct {
	Time      time.Time
	ProbeData []float64
}

func (td ThermoworksData) GetProbeData(pos ProbePosition) float64 {
	return td.ProbeData[pos-1]
}

func (td ThermoworksData) appendProbeData(lineData []opts.LineData, pos ProbePosition) []opts.LineData {
	if pos == ProbePositionNone {
		return lineData
	}
	probeData := td.GetProbeData(pos)
	if probeData <= 0 {
		return lineData
	}

	return append(lineData, opts.LineData{
		Value: []any{td.Time.Format(time.RFC3339), probeData},
	})
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
