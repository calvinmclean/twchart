package twchart

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"regexp"
	"strings"
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

// func parseDuration(input string) (time.Duration, error) {

// }

func parseTime(input string, date time.Time) (time.Time, error) {
	var result time.Time
	if input[0] == '+' {
		d, err := time.ParseDuration(input[1:])
		if err != nil {
			return time.Time{}, err
		}
		result = date.Add(d)
	} else {
		var err error
		result, err = time.ParseInLocation(time.Kitchen, input, time.Local)
		if err != nil {
			return time.Time{}, err
		}
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
		result.Second(),
		result.Nanosecond(),
		time.Local,
	), nil
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

func ParseLine(in []byte, currentDate time.Time) (SessionPart, time.Time, error) {
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
		event.Time, err = parseTime(string(match[1]), currentDate)
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

	stageTime, err := parseTime(stageTimeStr, currentDate)
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
