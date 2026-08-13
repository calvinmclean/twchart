package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/calvinmclean/twchart"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"Zero", 0, "0s"},
		{"Seconds", 30 * time.Second, "30s"},
		{"Minutes", 3 * time.Minute, "3m"},
		{"MinutesAndSeconds", 2*time.Minute + 30*time.Second, "2m30s"},
		{"Hours", time.Hour, "1h"},
		{"HoursAndMinutes", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"HoursMinutesSeconds", time.Hour + 2*time.Minute + 3*time.Second, "1h2m3s"},
		{"Milliseconds", 500 * time.Millisecond, "500ms"},
		{"Negative", -3 * time.Minute, "-3m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDuration(tt.duration))
		})
	}
}

func TestEventRow(t *testing.T) {
	start := time.Date(2025, time.May, 24, 18, 50, 0, 0, time.Local)
	prev := time.Date(2025, time.May, 24, 18, 51, 0, 0, time.Local)
	event := time.Date(2025, time.May, 24, 18, 53, 0, 0, time.Local)

	t.Run("WithPreviousEvent", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		data := map[string]any{
			"Event":            twchart.Event{Note: "finished mixing biga", Time: event},
			"PrevEventTime":    prev,
			"SessionStartTime": start,
		}

		result := eventRow.Render(r, data)

		assert.Contains(t, result, "6:53PM")
		assert.Contains(t, result, "(+2m | 3m elapsed)")
	})

	t.Run("FirstEvent", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		data := map[string]any{
			"Event":            twchart.Event{Note: "preheat", Time: start},
			"PrevEventTime":    time.Time{},
			"SessionStartTime": start,
		}

		result := eventRow.Render(r, data)

		assert.Contains(t, result, "6:50PM")
		assert.Contains(t, result, "(+0s)")
		assert.NotContains(t, result, "elapsed")
	})

	t.Run("NoStartTime", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		data := map[string]any{
			"Event":            twchart.Event{Note: "no start", Time: event},
			"PrevEventTime":    prev,
			"SessionStartTime": time.Time{},
		}

		result := eventRow.Render(r, data)

		assert.Contains(t, result, "6:53PM")
		assert.Contains(t, result, "(+2m)")
		assert.NotContains(t, result, "elapsed")
	})
}
