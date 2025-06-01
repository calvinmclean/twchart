package thermoworksbread

import (
	"bytes"
	"encoding"
	"io"
	"time"
)

var _ encoding.TextUnmarshaler = &Session{}
var _ io.Writer = &Session{}

// Write writes data from p into the BreadData struct
func (s *Session) Write(p []byte) (int, error) {
	return len(p), s.UnmarshalText(p)
}

// UnmarshalText parses the input bytes into the BreadData struct
func (s *Session) UnmarshalText(input []byte) error {
	var currentDate time.Time
	for _, line := range bytes.Split(input, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		result, newCurrentDate, err := ParseLine([]byte(line), currentDate)
		if err != nil {
			return err
		}
		result.AddToSession(s)

		currentDate = newCurrentDate
	}

	return nil
}

func isNextDay(currentTime, newTime time.Time) bool {
	// if currentTime is PM and newTime is AM, it is the next day
	return currentTime.Hour() >= 12 && newTime.Hour() < 12
}

func parseTime(input string, date time.Time) (time.Time, error) {
	result, err := time.ParseInLocation(time.Kitchen, input, time.Local)
	if err != nil {
		return time.Time{}, err
	}

	if isNextDay(date, result) {
		date = date.AddDate(0, 0, 1)
	}

	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		result.Hour(),
		result.Minute(),
		0, 0, time.Local,
	), nil
}
