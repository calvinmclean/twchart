package thermoworksbread

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("ParseName", func(t *testing.T) {
		bd := &BreadData{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Ciabatta"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "Ciabatta", bd.Name)
	})

	t.Run("ParseProbePosition", func(t *testing.T) {
		bd := &BreadData{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Ambient probe: 1"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "Ambient", bd.Probes[0].Name)
		assert.Equal(t, ProbePosition(ProbePosition1), bd.Probes[0].Position)

		input = "Other probe: 2"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "Other", bd.Probes[1].Name)
		assert.Equal(t, ProbePosition(ProbePosition2), bd.Probes[1].Position)

	})

	t.Run("ParseNote", func(t *testing.T) {
		bd := &BreadData{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Note: 8:10PM: preparing to make biga"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "preparing to make biga", bd.Events[0].Note)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), bd.Events[0].Time)

		input = "Note: 8:10PM: preparing to make poolish"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "preparing to make poolish", bd.Events[1].Note)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), bd.Events[1].Time)
	})

	t.Run("ParseStage", func(t *testing.T) {
		bd := &BreadData{}
		currentDate := time.Date(2025, time.May, 1, 0, 0, 0, 0, time.Local)

		input := "Preferment: 8:10PM"
		result, currentDate, err := ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "Preferment", bd.Stages[0].Name)
		assert.Equal(t, time.Date(2025, time.May, 1, 20, 10, 0, 0, time.Local), bd.Stages[0].Start)

		input = "Bulk ferment: 8:10AM"
		result, currentDate, err = ParseLine([]byte(input), currentDate)
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, "Bulk ferment", bd.Stages[1].Name)
		assert.Equal(t, time.Date(2025, time.May, 2, 8, 10, 0, 0, time.Local), bd.Stages[1].Start)
		assert.Equal(t, time.Date(2025, time.May, 2, 8, 10, 0, 0, time.Local), bd.Stages[0].End)
		assert.Equal(t, 12*time.Hour, bd.Stages[0].Duration)

		t.Run("ParseDone", func(t *testing.T) {
			input := "done: 8:30AM"
			result, currentDate, err = ParseLine([]byte(input), currentDate)
			assert.NoError(t, err)
			result.AddToSession(bd)
			assert.Equal(t, time.Date(2025, time.May, 2, 8, 30, 0, 0, time.Local), bd.Stages[1].End)
			assert.Equal(t, 20*time.Minute, bd.Stages[1].Duration)
		})
	})

	t.Run("ParseDate", func(t *testing.T) {
		bd := &BreadData{}

		input := "Date: 2025-05-24"
		result, currentDate, err := ParseLine([]byte(input), time.Time{})
		assert.NoError(t, err)
		result.AddToSession(bd)
		assert.Equal(t, time.Date(2025, time.May, 24, 0, 0, 0, 0, time.Local), bd.Date)
		assert.Equal(t, time.Date(2025, time.May, 24, 0, 0, 0, 0, time.Local), currentDate)
	})
}
