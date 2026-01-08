package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/twchart"
	"github.com/calvinmclean/twchart/storage"
	"github.com/calvinmclean/twchart/storage/db"
	"github.com/rs/xid"
)

type storageAdapter struct {
	*storage.Client
}

var _ babyapi.Storage[*SessionResource] = storageAdapter{}

// Convert database models to API resource
func (c storageAdapter) dbSessionToAPIResource(
	session db.Session,
	probes []db.Probe,
	stages []db.Stage,
	events []db.Event,
	thermoworksData []db.ThermoworksDatum,
) (*SessionResource, error) {
	resource := &SessionResource{
		Session: twchart.Session{
			Name:       session.Name,
			Date:       session.Date,
			StartTime:  session.StartTime.Time,
			UploadedAt: session.UploadedAt,
		},
	}
	// Convert string ID to xid.ID for the DefaultResource
	xidID, err := xid.FromString(session.ID)
	if err != nil {
		return nil, fmt.Errorf("error converting session ID: %w", err)
	}
	resource.Session.ID.ID = xidID

	// Convert probes
	for _, probe := range probes {
		resource.Session.Probes = append(resource.Session.Probes, twchart.Probe{
			Name:     probe.Name,
			Position: twchart.ProbePosition(probe.Position),
		})
	}

	// Convert stages
	for _, stage := range stages {
		s := twchart.Stage{
			Name:  stage.Name,
			Start: stage.Start,
			End:   stage.End.Time,
		}
		if stage.Duration.Valid {
			s.Duration = time.Duration(stage.Duration.Int64)
		}
		resource.Session.Stages = append(resource.Session.Stages, s)
	}

	// Convert events
	for _, event := range events {
		resource.Session.Events = append(resource.Session.Events, twchart.Event{
			Note: event.Note,
			Time: event.Time,
		})
	}

	resource.Session.Data = thermoworksDataFromDB(thermoworksData)

	return resource, nil
}

func thermoworksDataFromDB(thermoworksData []db.ThermoworksDatum) []twchart.ThermoworksData {
	out := []twchart.ThermoworksData{}
	for _, data := range thermoworksData {
		td := twchart.ThermoworksData{
			Time: data.Timestamp,
		}

		probeData := make([]float64, 6)
		if data.Probe1Temp.Valid {
			probeData[0] = data.Probe1Temp.Float64
		}
		if data.Probe2Temp.Valid {
			probeData[1] = data.Probe2Temp.Float64
		}
		if data.Probe3Temp.Valid {
			probeData[2] = data.Probe3Temp.Float64
		}
		if data.Probe4Temp.Valid {
			probeData[3] = data.Probe4Temp.Float64
		}
		if data.Probe5Temp.Valid {
			probeData[4] = data.Probe5Temp.Float64
		}
		if data.Probe6Temp.Valid {
			probeData[5] = data.Probe6Temp.Float64
		}
		td.ProbeData = probeData

		out = append(out, td)
	}
	return out
}

func (c storageAdapter) Get(ctx context.Context, id string) (*SessionResource, error) {
	session, err := c.Queries.GetSession(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, babyapi.ErrNotFound
		}
		return nil, fmt.Errorf("error getting session: %w", err)
	}

	probes, err := c.Queries.GetProbesBySession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting probes: %w", err)
	}

	stages, err := c.Queries.GetStagesBySession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting stages: %w", err)
	}

	events, err := c.Queries.GetEventsBySession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %w", err)
	}

	resource, err := c.dbSessionToAPIResource(session, probes, stages, events, nil)
	if err != nil {
		return nil, fmt.Errorf("error converting session to API resource: %w", err)
	}
	return resource, nil
}

func (c storageAdapter) Search(ctx context.Context, parentID string, query url.Values) ([]*SessionResource, error) {
	sessions, err := c.Queries.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing sessions: %w", err)
	}

	var resources []*SessionResource
	for _, session := range sessions {
		sessionResource, err := c.Get(ctx, session.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting session: %w", err)
		}

		resources = append(resources, sessionResource)
	}

	return resources, nil
}

func (c storageAdapter) Set(ctx context.Context, sessionResource *SessionResource) error {
	sessionID := string(sessionResource.GetID())

	// Check if session exists
	_, err := c.Queries.GetSession(ctx, sessionID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing session: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create new session
		_, err = c.Queries.CreateSession(ctx, db.CreateSessionParams{
			ID:         sessionID,
			Name:       sessionResource.Session.Name,
			Date:       sessionResource.Session.Date,
			StartTime:  sql.NullTime{Time: sessionResource.Session.StartTime, Valid: !sessionResource.Session.StartTime.IsZero()},
			UploadedAt: sessionResource.Session.UploadedAt,
		})
		if err != nil {
			return fmt.Errorf("error creating session: %w", err)
		}
	} else {
		// Update existing session
		_, err = c.Queries.UpdateSession(ctx, db.UpdateSessionParams{
			Name:      sessionResource.Session.Name,
			Date:      sessionResource.Session.Date,
			StartTime: sql.NullTime{Time: sessionResource.Session.StartTime, Valid: !sessionResource.Session.StartTime.IsZero()},
			ID:        sessionID,
		})
		if err != nil {
			return fmt.Errorf("error updating session: %w", err)
		}

		// Delete existing related data
		err = c.Queries.DeleteProbesBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("error deleting existing probes: %w", err)
		}
		err = c.Queries.DeleteStagesBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("error deleting existing stages: %w", err)
		}
		err = c.Queries.DeleteEventsBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("error deleting existing events: %w", err)
		}
		err = c.Queries.DeleteThermoworksDataBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("error deleting existing thermoworks data: %w", err)
		}
	}

	// Insert probes
	for _, probe := range sessionResource.Session.Probes {
		_, err = c.Queries.CreateProbe(ctx, db.CreateProbeParams{
			SessionID: sessionID,
			Name:      probe.Name,
			Position:  int64(probe.Position),
		})
		if err != nil {
			return fmt.Errorf("error creating probe: %w", err)
		}
	}

	// Insert stages
	for _, stage := range sessionResource.Session.Stages {
		_, err = c.Queries.CreateStage(ctx, db.CreateStageParams{
			SessionID: sessionID,
			Name:      stage.Name,
			Start:     stage.Start,
			End:       sql.NullTime{Time: stage.End, Valid: !stage.End.IsZero()},
			Duration:  sql.NullInt64{Int64: int64(stage.Duration), Valid: stage.Duration != 0},
		})
		if err != nil {
			return fmt.Errorf("error creating stage: %w", err)
		}
	}

	// Insert events
	for _, event := range sessionResource.Session.Events {
		_, err = c.Queries.CreateEvent(ctx, db.CreateEventParams{
			SessionID: sessionID,
			Note:      event.Note,
			Time:      event.Time,
		})
		if err != nil {
			return fmt.Errorf("error creating event: %w", err)
		}
	}

	// Insert thermoworks data
	err = c.storeThermoworksData(ctx, sessionID, sessionResource.Session.Data)
	if err != nil {
		return err
	}

	return nil
}

func (c storageAdapter) storeThermoworksData(ctx context.Context, sessionID string, data []twchart.ThermoworksData) error {
	for _, data := range data {
		probeData := make([]sql.NullFloat64, 6)
		for i, temp := range data.ProbeData {
			if i < len(probeData) && temp > 0 {
				probeData[i] = sql.NullFloat64{Float64: temp, Valid: true}
			}
		}

		_, err := c.Queries.CreateThermoworksData(ctx, db.CreateThermoworksDataParams{
			SessionID:  sessionID,
			Timestamp:  data.Time,
			Probe1Temp: probeData[0],
			Probe2Temp: probeData[1],
			Probe3Temp: probeData[2],
			Probe4Temp: probeData[3],
			Probe5Temp: probeData[4],
			Probe6Temp: probeData[5],
		})
		if err != nil {
			return fmt.Errorf("error creating thermoworks data: %w", err)
		}
	}

	return nil
}

func (c storageAdapter) Delete(ctx context.Context, id string) error {
	err := c.Queries.DeleteSession(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}
	return nil
}
