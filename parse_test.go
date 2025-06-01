package thermoworksbread

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
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
