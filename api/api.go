package api

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/calvinmclean/twchart"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/extensions"
	"github.com/go-chi/render"
)

type sessionResource struct {
	babyapi.DefaultResource
	Session    twchart.Session
	UploadedAt time.Time
}

var _ babyapi.HTMLer = &sessionResource{}

func (s sessionResource) HTML(w http.ResponseWriter, r *http.Request) string {
	return sessionDetail.Render(r, s)
}

func (s *sessionResource) Bind(r *http.Request) error {
	if s.UploadedAt.Equal(time.Time{}) {
		s.UploadedAt = time.Now()
	}

	err := s.DefaultResource.Bind(r)
	if err != nil {
		return err
	}
	return nil
}

var _ babyapi.Patcher[*sessionResource] = &sessionResource{}

func (s *sessionResource) Patch(newSession *sessionResource) *babyapi.ErrResponse {
	if !newSession.Session.StartTime.IsZero() && s.Session.StartTime.IsZero() {
		s.Session.StartTime = newSession.Session.StartTime
	}
	return nil
}

type API struct {
	*babyapi.API[*sessionResource]
}

func New() API {
	api := API{}
	api.API = babyapi.NewAPI("Sessions", "/sessions", func() *sessionResource { return &sessionResource{} })
	api.API.AddCustomRootRoute(http.MethodGet, "/", http.RedirectHandler("/sessions", http.StatusFound))
	api.API.AddCustomRoute(http.MethodPost, "/upload-csv", babyapi.Handler(api.loadCSVToLatestSession))
	api.SetSearchResponseWrapper(func(sr []*sessionResource) render.Renderer {
		return allSessionsWrapper{ResourceList: babyapi.ResourceList[*sessionResource]{Items: sr}}
	})
	api.API.AddCustomIDRoute(http.MethodGet, "/chart", api.GetRequestedResourceAndDo(api.renderChart))
	api.API.AddCustomIDRoute(http.MethodPost, "/add-event", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.Event](api)))
	api.API.AddCustomIDRoute(http.MethodPost, "/add-stage", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.Stage](api)))
	api.API.AddCustomIDRoute(http.MethodPost, "/done", api.GetRequestedResourceAndDo(sessionPartHandler[twchart.DoneTime](api)))

	// Use custom text unmarshalling/decoding for Sessions
	render.Decode = func(r *http.Request, v any) error {
		if render.GetRequestContentType(r) != render.ContentTypePlainText {
			return render.DefaultDecoder(r, v)
		}

		sessionTarget, ok := v.(*sessionResource)
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

func sessionPartHandler[T twchart.SessionPart](a API) func(http.ResponseWriter, *http.Request, *sessionResource) (render.Renderer, *babyapi.ErrResponse) {
	return func(_ http.ResponseWriter, r *http.Request, sr *sessionResource) (render.Renderer, *babyapi.ErrResponse) {
		var sessionPart T
		if err := render.DefaultDecoder(r, &sessionPart); err != nil {
			return nil, babyapi.ErrInvalidRequest(fmt.Errorf("error parsing SessionPart: %w", err))
		}

		sessionPart.AddToSession(&sr.Session)

		err := a.Storage.Set(r.Context(), sr)
		if err != nil {
			return nil, babyapi.InternalServerError(err)
		}
		return nil, nil
	}
}

func (a *API) loadCSVToLatestSession(w http.ResponseWriter, r *http.Request) render.Renderer {
	// TODO: sessions can be pretty large. I might want to create a way to do this with an iterator or other more memory-efficient option
	// An iterator would require babyapi changes but could be a good improvement
	allSessions, err := a.API.Storage.Search(r.Context(), "", nil)
	if err != nil {
		return babyapi.InternalServerError(err)
	}

	latest := &sessionResource{UploadedAt: time.Time{}}
	for _, s := range allSessions {
		if s.UploadedAt.After(latest.UploadedAt) {
			latest = s
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/csv" {
		return babyapi.ErrInvalidRequest(fmt.Errorf("unexpected Content-Type: %s", contentType))
	}

	err = latest.Session.LoadData(r.Body)
	if err != nil {
		return babyapi.ErrInvalidRequest(err)
	}

	err = a.API.Storage.Set(r.Context(), latest)
	if err != nil {
		return babyapi.InternalServerError(err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a *API) Setup(storeFilename string) {
	a.API.ApplyExtension(extensions.KeyValueStorage[*sessionResource]{
		KVConnectionConfig: extensions.KVConnectionConfig{Filename: storeFilename},
	})
}

func (*API) renderChart(w http.ResponseWriter, r *http.Request, sr *sessionResource) (render.Renderer, *babyapi.ErrResponse) {
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
