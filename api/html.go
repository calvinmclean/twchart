package api

import (
	"net/http"
	"slices"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/html"
)

const (
	sessionDetail         = html.Template("sessionDetail")
	sessionDetailTemplate = `{{ define "sessionDetail" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{ .Session.Name }} – {{ .Session.Date.Format "2006-01-02" }}</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/css/uikit.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit-icons.min.js"></script>
</head>
<body class="uk-background-muted uk-padding">

    <div class="uk-container uk-container-small">
        <!-- Header -->
        <div class="uk-flex uk-flex-between uk-flex-middle">
            <h1 class="uk-heading-line"><span>{{ .Session.Name }}</span></h1>
            <a href="/twcharts/{{ .DefaultResource.ID }}/chart" class="uk-button uk-button-default uk-button-small">Chart</a>
        </div>
        <p class="uk-text-meta">{{ .Session.Date.Format "Monday, Jan 2, 2006" }}</p>

        <!-- Stages -->
        {{ if .Session.Stages }}
        <div class="uk-card uk-card-default uk-card-body uk-margin">
            <h3 class="uk-card-title">Stages</h3>
            <table class="uk-table uk-table-divider uk-table-small">
                <thead>
                    <tr>
                        <th>Stage</th>
                        <th>Start</th>
                        <th>End</th>
                        <th>Duration</th>
                    </tr>
                </thead>
                <tbody>
                {{ range .Session.Stages }}
                    <tr>
                        <td>{{ .Name }}</td>
                        <td>{{ .Start.Format "3:04PM" }}</td>
                        <td>{{ if not .End.IsZero }}{{ .End.Format "3:04PM" }}{{ else }}–{{ end }}</td>
                        <td>{{ if .Duration }}{{ .Duration }}{{ else }}–{{ end }}</td>
                    </tr>
                {{ end }}
                </tbody>
            </table>
        </div>
        {{ end }}

        <!-- Events -->
        {{ if .Session.Events }}
        <div class="uk-card uk-card-default uk-card-body uk-margin">
            <h3 class="uk-card-title">Notes</h3>
            <ul class="uk-list uk-list-striped">
                {{ range .Session.Events }}
                <li class="uk-flex uk-flex-between">
                    <span>{{ .Note }}</span>
                    <span class="uk-text-meta">{{ .Time.Format "3:04PM" }}</span>
                </li>
                {{ end }}
            </ul>
        </div>
        {{ end }}

        <!-- Probes -->
        {{ if .Session.Probes }}
        <div class="uk-card uk-card-default uk-card-body uk-margin">
            <h3 class="uk-card-title">Probes</h3>
            <ul class="uk-subnav uk-subnav-divider">
                {{ range .Session.Probes }}
                <li><strong>{{ .Name }}</strong>: {{ .Position }}</li>
                {{ end }}
            </ul>
        </div>
        {{ end }}
    </div>
</body>
</html>
{{ end }}`

	listSessions         = html.Template("listSessions")
	listSessionsTemplate = `{{ define "listSessions" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Sessions</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/css/uikit.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit-icons.min.js"></script>
</head>
<body class="uk-background-muted uk-padding">

    <div class="uk-container uk-container-small">

        <h1 class="uk-heading-line uk-text-center"><span>Sessions</span></h1>

        {{ if . }}
        <ul class="uk-list uk-list-divider uk-margin">
            {{ range . }}
            <li class="uk-flex uk-flex-between uk-flex-middle">
                <div>
                    <h3 class="uk-margin-remove">
                        <a class="uk-link-heading" href="/twcharts/{{ .DefaultResource.ID }}">{{ .Session.Name }}</a>
                    </h3>
                    <p class="uk-text-meta uk-margin-remove-top">
                        {{ .Session.Date.Format "Monday, Jan 2, 2006" }}
                    </p>
                </div>
                <div>
                    <a href="/twcharts/{{ .DefaultResource.ID }}/chart" class="uk-button uk-button-default uk-button-small">Chart</a>
                </div>
            </li>
            {{ end }}
        </ul>
        {{ else }}
        <p class="uk-text-center uk-text-muted">No sessions found.</p>
        {{ end }}

    </div>

</body>
</html>
{{ end }}`
)

func init() {
	html.SetMap(map[string]string{
		string(sessionDetail): sessionDetailTemplate,
		string(listSessions):  listSessionsTemplate,
	})
}

// allSessionsWrapper allows rendering an HTML page that lists all charts
type allSessionsWrapper struct {
	babyapi.ResourceList[*sessionResource]
}

func (as allSessionsWrapper) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (as allSessionsWrapper) HTML(_ http.ResponseWriter, r *http.Request) string {
	slices.SortFunc(as.Items, func(a, b *sessionResource) int {
		return a.Session.StartTime.Compare(b.Session.StartTime)
	})
	return listSessions.Render(r, as.Items)
}
