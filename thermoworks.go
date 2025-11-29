package twchart

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	ProbePositionNone = iota
	ProbePosition1
	ProbePosition2
	ProbePosition3
	ProbePosition4
	ProbePosition5
	probePositionInvalid
)

type ProbePosition uint

func (pp *ProbePosition) UnmarshalJSON(input []byte) error {
	return pp.UnmarshalText(input)
}

func (pp *ProbePosition) UnmarshalText(input []byte) error {
	val, err := strconv.Atoi(string(input))
	if err != nil {
		return err
	}

	if val < ProbePositionNone || val >= probePositionInvalid {
		return fmt.Errorf("invalid ProbePosition: %d", val)
	}

	*pp = ProbePosition(val)
	return nil
}

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
		return append(lineData, opts.LineData{
			Value: []any{td.Time.Format(time.RFC3339), nil},
		})
	}

	return append(lineData, opts.LineData{
		Value: []any{td.Time.Format(time.RFC3339), probeData},
	})
}

func iterCSV(reader *csv.Reader) (iter.Seq2[ThermoworksData, error], error) {
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	if len(headers) < 2 || headers[0] != "DateTime" {
		return nil, fmt.Errorf("unexpected header format")
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
	}, nil
}
