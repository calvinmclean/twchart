package api

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/calvinmclean/twchart"
	"github.com/calvinmclean/twchart/storage"
	"github.com/rs/xid"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/extensions"
	"github.com/go-chi/render"
)

type SessionResource struct {
	*babyapi.DefaultRenderer
	twchart.Session
}

func (s SessionResource) GetID() string {
	return s.Session.ID.String()
}

func (s SessionResource) ParentID() string {
	return ""
}

var _ babyapi.HTMLer = &SessionResource{}

func (s SessionResource) HTML(w http.ResponseWriter, r *http.Request) string {
	return sessionDetail.Render(r, s)
}

func (s *SessionResource) Bind(r *http.Request) error {
	if s.Session.UploadedAt.Equal(time.Time{}) {
		s.Session.UploadedAt = time.Now()
	}

	err := s.Session.ID.Bind(r)
	if err != nil {
		return err
	}
	return nil
}

var _ babyapi.Patcher[*SessionResource] = &SessionResource{}

func (s *SessionResource) Patch(newSession *SessionResource) *babyapi.ErrResponse {
	if !newSession.Session.StartTime.IsZero() && s.Session.StartTime.IsZero() {
		s.Session.StartTime = newSession.Session.StartTime
	}
	return nil
}

type API struct {
	*babyapi.API[*SessionResource]

	// this currently won't work correctly for multiple users viewing the same session, but that's fine
	sseChans map[string]chan *babyapi.ServerSentEvent

	storageAdapter storageAdapter
}

// NewSessionResource creates a sessionResource from a Session
func NewSessionResource(session twchart.Session) *SessionResource {
	return &SessionResource{
		Session: session,
	}
}

func New() *API {
	api := &API{
		sseChans: map[string]chan *babyapi.ServerSentEvent{},
	}
	api.API = babyapi.NewAPI("Sessions", "/sessions", func() *SessionResource { return &SessionResource{} })
	api.API.AddCustomRootRoute(http.MethodGet, "/", http.RedirectHandler("/sessions", http.StatusFound))
	api.API.AddCustomRoute(http.MethodPost, "/upload-csv", babyapi.Handler(api.loadCSVToLatestSession))
	api.SetSearchResponseWrapper(func(sr []*SessionResource) render.Renderer {
		return allSessionsWrapper{ResourceList: babyapi.ResourceList[*SessionResource]{Items: sr}}
	})
	api.API.AddCustomIDRoute(http.MethodGet, "/chart", api.GetRequestedResourceAndDo(api.renderChart))
	api.API.AddCustomIDRoute(http.MethodPost, "/add-event", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.Event](api)))
	api.API.AddCustomIDRoute(http.MethodPost, "/add-stage", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.Stage](api)))
	api.API.AddCustomIDRoute(http.MethodPost, "/done", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.DoneTime](api)))
	api.API.AddCustomIDRoute(http.MethodGet, "/updates", http.HandlerFunc(api.sseUpdateHandler))

	// Use custom text unmarshalling/decoding for Sessions
	render.Decode = func(r *http.Request, v any) error {
		if render.GetRequestContentType(r) != render.ContentTypePlainText {
			return render.DefaultDecoder(r, v)
		}

		sessionTarget, ok := v.(*SessionResource)
		if !ok {
			return fmt.Errorf("unsupported target for plaintext decoder: %T", v)
		}

		var s twchart.Session
		_, err := io.Copy(&s, r.Body)
		if err != nil {
			return fmt.Errorf("error parsing Session: %w", err)
		}

		sessionTarget.Session = s

		return nil
	}

	return api
}

func (a API) sseUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := a.API.GetIDParam(r)

	sseChan := make(chan *babyapi.ServerSentEvent)
	a.sseChans[id] = sseChan
	defer func() {
		close(sseChan)
		delete(a.sseChans, id)
	}()

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	for {
		select {
		case e := <-sseChan:
			e.Write(w)
		case <-r.Context().Done():
			return
		case <-a.Done():
			return
		}
	}
}

func sessionPartHandler[T twchart.SessionPart](a *API) func(http.ResponseWriter, *http.Request, *SessionResource) (render.Renderer, *babyapi.ErrResponse) {
	return func(w http.ResponseWriter, r *http.Request, sr *SessionResource) (render.Renderer, *babyapi.ErrResponse) {
		var sessionPart T
		if err := render.DefaultDecoder(r, &sessionPart); err != nil {
			return nil, babyapi.ErrInvalidRequest(fmt.Errorf("error parsing SessionPart: %w", err))
		}

		sessionPart.AddToSession(&sr.Session)

		err := a.Storage.Set(r.Context(), sr)
		if err != nil {
			return nil, babyapi.InternalServerError(err)
		}

		logger := babyapi.GetLoggerFromContext(r.Context())

		// use ServerSentEvents to provide live updates to the UI
		sseChan, ok := a.sseChans[sr.GetID()]
		if !ok {
			logger.Info("no listeners for server-sent event")
			return nil, nil
		}

		event := &babyapi.ServerSentEvent{}
		switch part := any(sessionPart).(type) {
		case twchart.Event:
			event.Event = "newSessionEvent"
			// Find previous event time for duration calculation
			var prevEventTime time.Time
			events := sr.Session.Events
			// The new event is already added, so the previous event is the second-to-last
			if len(events) > 1 {
				prevEventTime = events[len(events)-2].Time
			}
			event.Data = eventRow.Render(r, map[string]any{
				"Event":            part,
				"PrevEventTime":    prevEventTime,
				"SessionStartTime": sr.Session.StartTime,
			})
		case twchart.Stage:
			event.Event = "newSessionStage"
			event.Data = stageRow.Render(r, sessionPart)
		case twchart.DoneTime:
			return nil, nil
			// nothing to do here since we don't append a stage and instead mark the last as ended.
			// not worth the effort to do right now
		}

		select {
		case sseChan <- event:
		default:
			logger.Info("no listeners for server-sent event")
		}

		return nil, nil
	}
}

func (a *API) loadCSVToLatestSession(w http.ResponseWriter, r *http.Request) render.Renderer {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/csv" {
		return babyapi.ErrInvalidRequest(fmt.Errorf("unexpected Content-Type: %s", contentType))
	}

	useSQL := a.storageAdapter.Client != nil

	var session *SessionResource
	if useSQL {
		sessionIDStr, err := a.storageAdapter.GetLatestSessionID(r.Context())
		if err != nil {
			return babyapi.InternalServerError(err)
		}
		sessionID, err := xid.FromString(sessionIDStr)
		if err != nil {
			return babyapi.InternalServerError(err)
		}

		// no real session is needed here since we just store the TW data
		session = &SessionResource{Session: twchart.Session{ID: babyapi.ID{ID: sessionID}}}
	} else {
		allSessions, err := a.API.Storage.Search(r.Context(), "", nil)
		if err != nil {
			return babyapi.InternalServerError(err)
		}

		session = &SessionResource{Session: twchart.Session{UploadedAt: time.Time{}}}
		for _, s := range allSessions {
			if s.Session.UploadedAt.After(session.Session.UploadedAt) {
				session = s
			}
		}
	}

	err := session.LoadData(r.Body)
	if err != nil {
		return babyapi.ErrInvalidRequest(err)
	}

	if useSQL {
		err = a.storageAdapter.storeThermoworksData(r.Context(), session.GetID(), session.Data)
		if err != nil {
			return babyapi.ErrInvalidRequest(err)
		}
	} else {
		err = a.API.Storage.Set(r.Context(), session)
		if err != nil {
			return babyapi.InternalServerError(err)
		}
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a *API) Setup(storeFilename string) error {
	ext := filepath.Ext(storeFilename)
	switch ext {
	case ".json":
		a.API.ApplyExtension(extensions.KeyValueStorage[*SessionResource]{
			KVConnectionConfig: extensions.KVConnectionConfig{Filename: storeFilename},
		})
		return nil
	case ".sql", ".sqlite", ".db":
		storageClient, err := storage.New(storeFilename)
		if err != nil {
			log.Fatalf("error creating DB client: %v", err)
		}
		a.storageAdapter = storageAdapter{storageClient}
		a.API.Storage = a.storageAdapter
		return nil
	default:
		return fmt.Errorf("unexpected extension for store file: %s", ext)
	}

}

func (a *API) renderChart(w http.ResponseWriter, r *http.Request, sr *SessionResource) (render.Renderer, *babyapi.ErrResponse) {
	if len(sr.Data) == 0 && a.storageAdapter.Client != nil {
		thermoworksData, err := a.storageAdapter.Client.GetThermoworksDataBySession(r.Context(), sr.GetID())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, babyapi.InternalServerError(err)
		}
		sr.Data = thermoworksDataFromDB(thermoworksData)
	}

	chart, err := twchart.Session(sr.Session).Chart()
	if err != nil {
		return nil, babyapi.InternalServerError(err)
	}

	snippet := chart.RenderSnippet()

	return chartView.Renderer(struct {
		Element template.HTML
		Script  template.HTML
		Title   string
		BackURL string
	}{
		Element: template.HTML(snippet.Element),
		Script:  template.HTML(snippet.Script),
		Title:   sr.Session.Name,
		BackURL: fmt.Sprintf("/sessions/%s", sr.GetID()),
	}), nil
}
