package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/calvinmclean/twchart"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/extensions"
	"github.com/go-chi/render"
)

// sessionAlias prevents the UnmarshalText from being used for JSON
type sessionAlias twchart.Session

type sessionResource struct {
	babyapi.DefaultResource
	Session    sessionAlias
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

	// Use custom text unmarshalling/decoding for Sessions
	render.Decode = func(r *http.Request, v any) error {
		if render.GetRequestContentType(r) == render.ContentTypePlainText {
			sessionTarget, ok := v.(*sessionResource)
			if !ok {
				return fmt.Errorf("unsupported target for plaintext decoder: %T", v)
			}

			var s twchart.Session
			_, err := io.Copy(&s, r.Body)
			if err != nil {
				return fmt.Errorf("error parsing Session: %w", err)
			}

			sessionTarget.Session = sessionAlias(s)

			return nil
		}
		return render.DefaultDecoder(r, v)
	}

	return api
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

	session := twchart.Session(latest.Session)
	err = session.LoadData(r.Body)
	if err != nil {
		return babyapi.ErrInvalidRequest(err)
	}

	latest.Session = sessionAlias(session)
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

func (*API) renderChart(w http.ResponseWriter, _ *http.Request, sr *sessionResource) (render.Renderer, *babyapi.ErrResponse) {
	chart, err := twchart.Session(sr.Session).Chart()
	if err != nil {
		return nil, babyapi.InternalServerError(err)
	}

	err = chart.Render(w)
	if err != nil {
		return nil, babyapi.InternalServerError(err)
	}
	return nil, nil
}
