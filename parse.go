package thermoworksbread

import (
	"bufio"
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var _ encoding.TextUnmarshaler = &BreadData{}
var _ io.Writer = &BreadData{}

// Write writes data from p into the BreadData struct
func (bd *BreadData) Write(p []byte) (int, error) {
	return len(p), bd.UnmarshalText(p)
}

// UnmarshalText parses the input bytes into the BreadData struct
func (bd *BreadData) UnmarshalText(input []byte) error {
	var currentDate time.Time
	scanner := bufio.NewScanner(bytes.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if bd.Name == "" {
			bd.Name = line
			continue
		}

		parts := strings.SplitN(line, ":", 2)

		if len(parts) != 2 {
			return fmt.Errorf("unexpected format: %v", parts)
		}

		value := strings.TrimSpace(parts[1])

		var err error
		switch strings.ToLower(parts[0]) {
		case "date":
			currentDate, err = time.ParseInLocation(time.DateOnly, value, time.Local)
			if err != nil {
				return fmt.Errorf("error parsing date: %w", err)
			}
		case "note":
			parsedTime, note, err := parseNote(value, currentDate)
			if err != nil {
				return fmt.Errorf("error parsing note: %w", err)
			}
			bd.AddEvents(Event{Time: parsedTime, Note: note})
			currentDate = parsedTime
		case "preferment":
			currentDate, err = parseStage(currentDate, value, bd.StartPreferment)
			if err != nil {
				return fmt.Errorf("error parsing preferment: %w", err)
			}
		case "bulk ferment":
			currentDate, err = parseStage(currentDate, value, bd.StartBulkFerment)
			if err != nil {
				return fmt.Errorf("error parsing bulk ferment: %w", err)
			}
		case "final proof":
			currentDate, err = parseStage(currentDate, value, bd.StartFinalProof)
			if err != nil {
				return fmt.Errorf("error parsing final proof: %w", err)
			}
		case "bake":
			currentDate, err = parseStage(currentDate, value, bd.StartBake)
			if err != nil {
				return fmt.Errorf("error parsing bake: %w", err)
			}
		case "done":
			currentDate, err = parseStage(currentDate, value, bd.EndBake)
			if err != nil {
				return fmt.Errorf("error parsing bake: %w", err)
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("error scanning: %w", err)
	}

	return nil
}

func parseNote(value string, currentTime time.Time) (time.Time, string, error) {
	// Find the second colon in " 6:53PM: ..."
	second := false
	i := strings.IndexFunc(value, func(r rune) bool {
		if r != ':' {
			return false
		}
		if !second {
			second = true
			return false
		}
		return true
	})
	if i < 0 {
		fmt.Println(value)
		return time.Time{}, "", errors.New("note is missing expected number of ':'")
	}

	parsedTime, err := parseTime(strings.TrimSpace(value[0:i]), currentTime)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing time: %w", err)
	}

	if isNextDay(currentTime, parsedTime) {
		parsedTime = parsedTime.AddDate(0, 0, 1)
	}

	note := strings.TrimSpace(value[i+1:])
	return parsedTime, note, nil
}

func isNextDay(currentTime, newTime time.Time) bool {
	// if currentTime is PM and newTime is AM, it is the next day
	return currentTime.Hour() >= 12 && newTime.Hour() < 12
}

func parseStage(currentTime time.Time, value string, handle func(time.Time)) (time.Time, error) {
	parsedTime, err := parseTime(value, currentTime)
	if err != nil {
		return time.Time{}, err
	}
	if isNextDay(currentTime, parsedTime) {
		parsedTime = parsedTime.AddDate(0, 0, 1)
	}
	handle(parsedTime)
	return parsedTime, nil
}

func parseTime(input string, date time.Time) (time.Time, error) {
	result, err := time.ParseInLocation(time.Kitchen, input, time.Local)
	if err != nil {
		return time.Time{}, err
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
