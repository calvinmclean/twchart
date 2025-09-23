package api

import (
	"net/http"

	"github.com/calvinmclean/babyapi"
)

// allSessionsWrapper allows rendering an HTML page that lists all charts
type allSessionsWrapper struct {
	babyapi.ResourceList[*sessionResource]
}

func (as allSessionsWrapper) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// TODO: make this better
func (as allSessionsWrapper) HTML(_ http.ResponseWriter, r *http.Request) string {
	html := "<html><body><h1>All Sessions</h1><ul>"
	for _, res := range as.Items {
		// Each link points to /twcharts/{id}
		html += "<li><a href=\"/twcharts/" + res.GetID() + "\">" + res.Session.Name + "</a></li>"
	}
	html += "</ul></body></html>"
	return html
}
