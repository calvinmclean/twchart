package thermoworksbread

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type SessionPart interface {
	AddToSession(*Session)
}

type Probe struct {
	Name     string
	Position ProbePosition
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
	noteRE  = regexp.MustCompile(`(?i)^Note:\s+(?P<timestamp>\d{1,2}:\d{2}(?:AM|PM)):\s+(?P<note>.+)$`)
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
