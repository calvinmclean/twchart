package thermoworksbread

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type SessionPart interface {
	AddToSession(*BreadData)
}

type Probe struct {
	Name     string
	Position ProbePosition
}

func (p Probe) AddToSession(bd *BreadData) {
	bd.Probes = append(bd.Probes, p)
}

func (s Stage) AddToSession(bd *BreadData) {
	prevIdx := len(bd.Stages) - 1
	bd.Stages = append(bd.Stages, s)

	// Finish previous stage
	if prevIdx == -1 {
		return
	}
	bd.Stages[prevIdx].Finish(s.Start)
}

type DoneTime time.Time

func (dt DoneTime) AddToSession(bd *BreadData) {
	// Finish the last stage
	prevIdx := len(bd.Stages) - 1
	if prevIdx == -1 {
		return
	}

	bd.Stages[prevIdx].Finish(time.Time(dt))
}

type SessionDate time.Time

func (sd SessionDate) AddToSession(bd *BreadData) {
	bd.Date = time.Time(sd)
}

func (e Event) AddToSession(bd *BreadData) {
	bd.Events = append(bd.Events, e)
}

type SessionName string

func (sn SessionName) AddToSession(bd *BreadData) {
	bd.Name = string(sn)
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
