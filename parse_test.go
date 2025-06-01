package twchart

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	t.Run("ParseName", func(t *testing.T) {
		s := &Session{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Ciabatta"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "Ciabatta", s.Name)
	})

	t.Run("ParseProbePosition", func(t *testing.T) {
		s := &Session{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Ambient probe: 1"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "Ambient", s.Probes[0].Name)
		assert.Equal(t, ProbePosition(ProbePosition1), s.Probes[0].Position)

		input = "Other probe: 2"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "Other", s.Probes[1].Name)
		assert.Equal(t, ProbePosition(ProbePosition2), s.Probes[1].Position)

	})

	t.Run("ParseNote", func(t *testing.T) {
		s := &Session{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Note: 8:10PM: preparing to make biga"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "preparing to make biga", s.Events[0].Note)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), s.Events[0].Time)

		input = "Note: 8:10PM: preparing to make poolish"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "preparing to make poolish", s.Events[1].Note)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), s.Events[1].Time)
	})

	t.Run("ParseStage", func(t *testing.T) {
		s := &Session{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Preferment: 8:10PM"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "Preferment", s.Stages[0].Name)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), s.Stages[0].Start)

		input = "Bulk ferment: 8:10AM"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, "Bulk ferment", s.Stages[1].Name)
		assert.Equal(t, time.Date(2025, time.May, 2, 8, 10, 0, 0, time.Local), s.Stages[1].Start)
		assert.Equal(t, time.Date(2025, time.May, 2, 8, 10, 0, 0, time.Local), s.Stages[0].End)
		assert.Equal(t, 12*time.Hour, s.Stages[0].Duration)

		t.Run("ParseDone", func(t *testing.T) {
			input := "done: 8:30AM"
			result, currentDate, err = ParseLine([]byte(input), currentDate)
			assert.NoError(t, err)
			result.AddToSession(s)
			assert.Equal(t, time.Date(2025, time.May, 2, 8, 30, 0, 0, time.Local), s.Stages[1].End)
			assert.Equal(t, 20*time.Minute, s.Stages[1].Duration)
		})
	})

	t.Run("ParseDate", func(t *testing.T) {
		s := &Session{}

		input := "Date: 2025-05-24"
		result, currentDate, err := ParseLine([]byte(input), time.Time{})
		assert.NoError(t, err)
		result.AddToSession(s)
		assert.Equal(t, time.Date(2025, time.May, 24, 0, 0, 0, 0, time.Local), s.Date)
		assert.Equal(t, time.Date(2025, time.May, 24, 0, 0, 0, 0, time.Local), currentDate)
	})
}

func TestParse(t *testing.T) {
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

	var bd Session
	// err := s.UnmarshalText([]byte(input))
	_, err := io.Copy(&bd, bytes.NewReader([]byte(input)))
	assert.NoError(t, err)
	assert.Equal(t, Session{
		Name: "Ciabatta",
		Date: time.Date(2025, time.May, 24, 0, 0, 0, 0, time.Local),
		Stages: []Stage{
			{
				Name:     "Preferment",
				Start:    time.Date(2025, time.May, 24, 18, 51, 0, 0, time.Local),
				End:      time.Date(2025, time.May, 25, 7, 00, 0, 0, time.Local),
				Duration: 12*time.Hour + 9*time.Minute,
			},
			{
				Name:     "Bulk ferment",
				Start:    time.Date(2025, time.May, 25, 7, 00, 0, 0, time.Local),
				End:      time.Date(2025, time.May, 25, 9, 00, 0, 0, time.Local),
				Duration: 2 * time.Hour,
			},
			{
				Name:     "Final Proof",
				Start:    time.Date(2025, time.May, 25, 9, 00, 0, 0, time.Local),
				End:      time.Date(2025, time.May, 25, 10, 30, 0, 0, time.Local),
				Duration: 90 * time.Minute,
			},
			{
				Name:     "Bake",
				Start:    time.Date(2025, time.May, 25, 10, 30, 0, 0, time.Local),
				End:      time.Date(2025, time.May, 25, 10, 55, 0, 0, time.Local),
				Duration: 25 * time.Minute,
			},
		},
		Events: []Event{
			{Note: "preparing to make biga", Time: time.Date(2025, time.May, 24, 18, 50, 0, 0, time.Local)},
			{Note: "finished mixing biga", Time: time.Date(2025, time.May, 24, 18, 53, 0, 0, time.Local)},
			{Note: "10 stretch and folds", Time: time.Date(2025, time.May, 25, 8, 0, 0, 0, time.Local)},
			{Note: "shaped dough", Time: time.Date(2025, time.May, 25, 9, 0, 0, 0, time.Local)},
			{Note: "bread is delicious and crunchy", Time: time.Date(2025, time.May, 25, 12, 0, 0, 0, time.Local)},
		},
		Probes: []Probe{
			{Name: "Ambient", Position: ProbePosition1},
			{Name: "Oven", Position: ProbePosition2},
			{Name: "Dough", Position: ProbePosition3},
			{Name: "Other", Position: ProbePosition4},
		},
	}, bd)
}
