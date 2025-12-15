package twchart

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var _ io.Writer = &Session{}

// Write writes data from p into the Session struct
func (s *Session) Write(p []byte) (int, error) {
	return len(p), s.FromText(p)
}

// FromText parses the input bytes into the Session struct
func (s *Session) FromText(input []byte) error {
	var currentDate time.Time
	for line := range bytes.SplitSeq(input, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		result, newCurrentDate, err := ParseLine([]byte(line), currentDate, s.StartTime)
		if err != nil {
			return err
		}
		result.AddToSession(s)

		currentDate = newCurrentDate

		// Set the session's StartTime for the first Stage or Event
		if s.StartTime.Equal(time.Time{}) {
			switch result.(type) {
			case Stage, Event:
				s.StartTime = currentDate
			}
		}
	}

	return nil
}

func isNextDay(currentTime, newTime time.Time) bool {
	// it is not the next day if this is the first time measurement
	if currentTime.Hour() == 0 && currentTime.Minute() == 0 && currentTime.Second() == 0 {
		return false
	}
	if newTime.Equal(currentTime) {
		return false
	}
	// if currentTime is PM and newTime is AM, it is the next day
	morningIsOver := currentTime.Hour() >= 12 && newTime.Hour() < 12
	if morningIsOver {
		return true
	}
	// Otherwise, if the new time is before the previous, nearly 24 hours have passed
	// This won't work for exactly 24 hours
	return newTime.Before(currentTime)
}

func parseTime(input string, date, startTime time.Time) (time.Time, error) {
	var result time.Time

	durationStr := strings.TrimPrefix(input, "+")
	d, err := time.ParseDuration(durationStr)
	if err != nil {
		result, err = parseTimestamp(input, date)
		if err != nil {
			return time.Time{}, err
		}
	} else if input[0] == '+' {
		// if it starts with +, add to previous time
		result = date.Add(d)
	} else {
		// Otherwise, add to start time
		result = startTime.Add(d)
	}

	if isNextDay(date, result) {
		result = result.AddDate(0, 0, 1)
	}

	return result, nil
}

// parseTimestamp will attempt to parse time.Kitchen or a combination of time.DateOnly + time.Kitchen
func parseTimestamp(input string, currentDate time.Time) (time.Time, error) {
	result, kitchenErr := time.ParseInLocation(time.Kitchen, input, time.Local)
	if kitchenErr == nil {
		// Kitchen time defaults to 0000-01-01 for the date part, so we add the current date (-1 on month and day)
		result = result.AddDate(currentDate.Year(), int(currentDate.Month())-1, currentDate.Day()-1)
		return result, nil
	}

	result, err := time.ParseInLocation(time.DateOnly+" "+time.Kitchen, input, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %w", errors.Join(kitchenErr, err))
	}
	return result, nil
}

// SessionPart is an interface that allows any parsed type to be applied to a Session
type SessionPart interface {
	AddToSession(*Session)
}

func (p Probe) AddToSession(s *Session) {
	s.Probes = append(s.Probes, p)
}

func (s Stage) AddToSession(session *Session) {
	prevIdx := len(session.Stages) - 1
	session.Stages = append(session.Stages, s)

	// Finish previous stage
	if prevIdx == -1 {
		return
	}
	session.Stages[prevIdx].Finish(s.Start)
}

type DoneTime time.Time

func (dt *DoneTime) UnmarshalJSON(in []byte) error {
	var d struct {
		Time time.Time `json:"time"`
	}
	err := json.Unmarshal(in, &d)
	if err != nil {
		return err
	}

	if d.Time.IsZero() {
		d.Time = time.Now()
	}

	*dt = DoneTime(d.Time)
	return nil
}

func (dt DoneTime) AddToSession(s *Session) {
	// Finish the last stage
	prevIdx := len(s.Stages) - 1
	if prevIdx == -1 {
		return
	}

	s.Stages[prevIdx].Finish(time.Time(dt))
}

type SessionDate time.Time

func (sd SessionDate) AddToSession(s *Session) {
	s.Date = time.Time(sd)
}

func (e Event) AddToSession(s *Session) {
	s.Events = append(s.Events, e)
}

type SessionName string

func (sn SessionName) AddToSession(s *Session) {
	s.Name = string(sn)
}

var (
	probeRE = regexp.MustCompile(`(?i)(?P<name>.+?)\s+probe:\s+(?P<number>\d+)`)
	noteRE  = regexp.MustCompile(`(?i)^Note:\s+(?P<timestamp>.+?):\s+(?P<note>.+)$`)
)

func ParseLine(in []byte, currentDate, startTime time.Time) (SessionPart, time.Time, error) {
	if !bytes.Contains(in, []byte{':'}) {
		return SessionName(in), currentDate, nil
	}

	if match := probeRE.FindSubmatch(in); len(match) == 3 {
		probe := Probe{
			Name: string(match[1]),
		}
		err := probe.Position.UnmarshalText(match[2])
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("error parsing ProbePosition %q: %w", string(match[2]), err)
		}

		return probe, currentDate, nil
	} else if match := noteRE.FindSubmatch(in); len(match) == 3 {
		event := Event{
			Note: string(match[2]),
		}
		var err error
		event.Time, err = parseTime(string(match[1]), currentDate, startTime)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("error parsing Note time %q: %w", string(match[1]), err)
		}

		return event, event.Time, nil
	}

	split := bytes.SplitN(in, []byte{':'}, 2)
	if len(split) != 2 {
		return nil, time.Time{}, fmt.Errorf("invalid input: %q", string(in))
	}

	stageName := strings.TrimSpace(string(split[0]))
	stageTimeStr := strings.TrimSpace(string(split[1]))

	if strings.ToLower(stageName) == "date" {
		date, err := time.ParseInLocation(time.DateOnly, stageTimeStr, time.Local)
		if err != nil {
			return nil, currentDate, fmt.Errorf("error parsing date: %w", err)
		}
		return SessionDate(date), date, nil
	}

	stageTime, err := parseTime(stageTimeStr, currentDate, startTime)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("error parsing Stage time %q: %w", stageTimeStr, err)
	}

	if strings.ToLower(stageName) == "done" {
		return DoneTime(stageTime), stageTime, nil
	}

	return Stage{
		Name:  stageName,
		Start: stageTime,
	}, stageTime, nil
}
