package thermoworksbread

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseBreadData(t *testing.T) {
	input := `Ciabatta
Date: 2025-05-24

Ambient Probe: 1
Oven Probe: 2
Dough Probe: 3
Other Probe: 4

Note: 6:50PM: preparing to make biga

Preferment: 6:51PM
Note: 6:53PM: finished mixing biga

Bulk ferment: 7:00AM
Note: 8:00AM: 10 stretch and folds

Final Proof: 9:00AM
Note: 9:00AM: shaped dough

Bake: 10:30AM
Done: 10:55AM

Note: 12:00PM: bread is delicious and crunchy
`

	var bd BreadData
	// err := bd.UnmarshalText([]byte(input))
	_, err := io.Copy(&bd, bytes.NewReader([]byte(input)))
	assert.NoError(t, err)
	assert.Equal(t, BreadData{
		Name: "Ciabatta",
		Preferment: Stage{
			Name:     "Preferment",
			Start:    time.Date(2025, time.May, 24, 18, 51, 0, 0, time.Local),
			End:      time.Date(2025, time.May, 25, 7, 00, 0, 0, time.Local),
			Duration: 12*time.Hour + 9*time.Minute,
		},
		BulkFerment: Stage{
			Name:     "Bulk Ferment",
			Start:    time.Date(2025, time.May, 25, 7, 00, 0, 0, time.Local),
			End:      time.Date(2025, time.May, 25, 9, 00, 0, 0, time.Local),
			Duration: 2 * time.Hour,
		},
		FinalProof: Stage{
			Name:     "Final Proof",
			Start:    time.Date(2025, time.May, 25, 9, 00, 0, 0, time.Local),
			End:      time.Date(2025, time.May, 25, 10, 30, 0, 0, time.Local),
			Duration: 90 * time.Minute,
		},
		Bake: Stage{
			Name:     "Bake",
			Start:    time.Date(2025, time.May, 25, 10, 30, 0, 0, time.Local),
			End:      time.Date(2025, time.May, 25, 10, 55, 0, 0, time.Local),
			Duration: 25 * time.Minute,
		},
		Events: []Event{
			{Note: "preparing to make biga", Time: time.Date(2025, time.May, 24, 18, 50, 0, 0, time.Local)},
			{Note: "finished mixing biga", Time: time.Date(2025, time.May, 24, 18, 53, 0, 0, time.Local)},
			{Note: "10 stretch and folds", Time: time.Date(2025, time.May, 25, 8, 0, 0, 0, time.Local)},
			{Note: "shaped dough", Time: time.Date(2025, time.May, 25, 9, 0, 0, 0, time.Local)},
			{Note: "bread is delicious and crunchy", Time: time.Date(2025, time.May, 25, 12, 0, 0, 0, time.Local)},
		},
		AmbientProbePosition: ProbePosition1,
		OvenProbePosition:    ProbePosition2,
		DoughProbePosition:   ProbePosition3,
		OtherProbePosition:   ProbePosition4,
	}, bd)
}
