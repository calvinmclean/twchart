package api

import (
	"net/http"
	"slices"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/html"
)

// getUniqueTypes extracts unique non-empty type values from sessions
func getUniqueTypes(sessions []*SessionResource) []string {
	typeMap := make(map[string]bool)
	for _, s := range sessions {
		if s.Session.Type != "" {
			typeMap[string(s.Session.Type)] = true
		}
	}
	types := make([]string, 0, len(typeMap))
	for t := range typeMap {
		types = append(types, t)
	}
	slices.Sort(types)
	return types
}

const (
	listSessions         = html.Template("listSessions")
	listSessionsTemplate = `{{ define "listSessions" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Sessions</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/css/uikit.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit-icons.min.js"></script>
</head>
<body class="uk-background-muted uk-padding">
    <div class="uk-container uk-container-small">
	    <ul class="uk-breadcrumb uk-margin-small-top">
		    <li><a href="/sessions">Sessions</a></li>
		    <li><span></span></li>
		</ul>

		<div class="uk-flex uk-flex-between uk-flex-middle">
            <h1 class="uk-heading-line"><span>Sessions</span></h1>
        </div>

        <!-- Type Filter -->
        <div class="uk-margin uk-flex uk-flex-middle">
            <select id="type-filter" class="uk-select uk-form-width-medium">
                <option value="">All Types</option>
                {{ $types := getUniqueTypes . }}
                {{ range $types }}
                <option value="{{ . }}">{{ . }}</option>
                {{ end }}
            </select>
        </div>

        {{ if . }}
        <ul id="sessions-list" class="uk-list uk-list-divider uk-margin">
            {{ range . }}
            <li class="uk-flex uk-flex-between uk-flex-middle session-item" data-type="{{ .Session.Type }}">
                <div>
                    <h3 class="uk-margin-remove">
                        <a class="uk-link-heading" href="/sessions/{{ .Session.ID }}">{{ .Session.Name }}</a>
                        {{ if .Session.Type }}<span class="uk-label uk-margin-small-left" style="background-color: #e5e5e5; color: #666;">{{ .Session.Type }}</span>{{ end }}
                    </h3>
                    <p class="uk-text-meta uk-margin-remove-top">
                        {{ .Session.Date.Format "Monday, Jan 2, 2006" }}
                    </p>
                </div>
                <div>
                    <a href="/sessions/{{ .Session.ID }}/chart" class="uk-button uk-button-default uk-button-small">Chart</a>
                </div>
            </li>
            {{ end }}
        </ul>
        {{ else }}
        <p class="uk-text-center uk-text-muted">No sessions found.</p>
        {{ end }}

    </div>

    <script>
        document.getElementById('type-filter').addEventListener('change', function() {
            var selectedType = this.value;
            var items = document.querySelectorAll('.session-item');
            var visibleCount = 0;

            items.forEach(function(item) {
                var itemType = item.getAttribute('data-type');
                if (selectedType === '' || itemType === selectedType) {
                    item.style.display = 'flex';
                    visibleCount++;
                } else {
                    item.style.display = 'none';
                }
            });

            // Show/hide the "No sessions found" message
            var noSessionsMsg = document.querySelector('.uk-text-center.uk-text-muted');
            if (noSessionsMsg) {
                noSessionsMsg.style.display = visibleCount === 0 ? 'block' : 'none';
            }
        });
    </script>

</body>
</html>
{{ end }}`

	stageRow         = html.Template("stageRow")
	stageRowTemplate = `<tr>
    <td>{{ .Name }}</td>
    <td>{{ .Start.Format "3:04PM" }}</td>
    <td>{{ if not .End.IsZero }}{{ .End.Format "3:04PM" }}{{ else }}–{{ end }}</td>
    <td>{{ if .Duration }}{{ .Duration }}{{ else }}–{{ end }}</td>
</tr>`

	eventRow         = html.Template("eventRow")
	eventRowTemplate = `<li class="uk-flex uk-flex-between">
    <span>{{ .Event.Note }}</span>
    <span class="uk-text-meta">
        {{ .Event.Time.Format "3:04PM" }}
        {{ if not (isZeroTime .PrevEventTime) }}
            <span class="uk-text-muted">(+{{ .Event.Time.Sub .PrevEventTime | formatDuration }})</span>
        {{ else if not (isZeroTime .SessionStartTime) }}
            <span class="uk-text-muted">(+{{ .Event.Time.Sub .SessionStartTime | formatDuration }})</span>
        {{ end }}
    </span>
</li>
`

	sessionDetail         = html.Template("sessionDetail")
	sessionDetailTemplate = `{{ define "sessionDetail" }}
<!DOCTYPE html>
<html lang="en">
<head>
   <meta charset="UTF-8">
   <meta name="viewport" content="width=device-width, initial-scale=1">
   <title>{{ .Session.Name }}</title>
   <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/css/uikit.min.css" />
   <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit.min.js"></script>
   <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit-icons.min.js"></script>

   <script src="https://unpkg.com/htmx.org@1.9.8"></script>
   <script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>
</head>
<body class="uk-background-muted uk-padding" hx-ext="sse" sse-connect="/sessions/{{ .Session.ID }}/updates">
   <div class="uk-container uk-container-small">
	    <ul class="uk-breadcrumb uk-margin-small-top">
	        <li><a href="/sessions">Sessions</a></li>
	        <li><span>{{ .Session.Name }}</span></li>
	    </ul>

       <!-- Header -->
       <div class="uk-flex uk-flex-between uk-flex-middle">
           <h1 class="uk-heading-line"><span>{{ .Session.Name }}</span></h1>
           <a href="/sessions/{{ .Session.ID }}/chart" class="uk-button uk-button-default uk-button-small">Chart</a>
       </div>
       <p class="uk-text-meta">{{ .Session.Date.Format "Monday, Jan 2, 2006" }}</p>

       <!-- Stages -->
       <div class="uk-card uk-card-default uk-card-body uk-margin">
           <h3 class="uk-card-title">Stages</h3>
           <div class="uk-overflow-auto">
	            <table class="uk-table uk-table-divider uk-table-small">
	                <thead>
	                    <tr>
	                        <th>Stage</th>
	                        <th>Start</th>
	                        <th>End</th>
	                        <th>Duration</th>
	                    </tr>
	                </thead>
	                <tbody sse-swap="newSessionStage" hx-swap="beforeend">
	                {{ range .Session.Stages }}
						{{ template "stageRow" . }}
	                {{ end }}
	                </tbody>
	            </table>
           </div>
       </div>

        <!-- Events -->
        <div class="uk-card uk-card-default uk-card-body uk-margin">
            <h3 class="uk-card-title">Notes</h3>
            <ul class="uk-list uk-list-striped" sse-swap="newSessionEvent" hx-swap="beforeend">
                {{ range $i, $e := .Session.Events }}
                    {{ $prevTime := zeroTime }}
                    {{ if gt $i 0 }}
                        {{ $prev := index $.Session.Events (sub $i 1) }}
                        {{ $prevTime = $prev.Time }}
                    {{ end }}
                    {{ template "eventRow" dict "Event" $e "PrevEventTime" $prevTime "SessionStartTime" $.Session.StartTime }}
                {{ end }}
            </ul>
        </div>

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

	chartView         = html.Template("chartView")
	chartViewTemplate = `{{ define "chartView" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .Title }} - Chart</title>

    <!-- UIkit -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/css/uikit.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/uikit@3.19.2/dist/js/uikit-icons.min.js"></script>

    <!-- Apache ECharts -->
    <script src="https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js"></script>
</head>
<body class="uk-background-muted uk-padding">
    <div class="uk-container uk-container-small">
	    <ul class="uk-breadcrumb uk-margin-small-top">
		    <li><a href="/sessions">Sessions</a></li>
		    <li><a href="{{ .BackURL }}">{{ .Title }}</a></li>
		    <li><span>Chart</span></li>
		</ul>
		<div class="uk-flex uk-flex-between uk-flex-middle">
            <h1 class="uk-heading-line"><span>{{ .Title }}</span></h1>

            <div class="uk-text-center uk-margin">
                <a href="{{ .BackURL }}" class="uk-button uk-button-default uk-button-small">← Back</a>
            </div>
        </div>
    </div>
    <div class="uk-container">
       	{{ .Element }}
        {{ .Script }}
    </div>
</body>
</html>
{{ end }}`
)

func init() {
	html.SetMap(map[string]string{
		string(sessionDetail): sessionDetailTemplate,
		string(listSessions):  listSessionsTemplate,
		string(chartView):     chartViewTemplate,
		string(stageRow):      stageRowTemplate,
		string(eventRow):      eventRowTemplate,
	})

	html.SetFuncs(func(r *http.Request) map[string]any {
		return map[string]any{
			"getUniqueTypes": getUniqueTypes,
			"sub": func(a, b int) int {
				return a - b
			},
			"dict": func(values ...any) map[string]any {
				if len(values)%2 != 0 {
					panic("dict requires an even number of arguments")
				}
				dict := make(map[string]any, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						panic("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict
			},
			"zeroTime": func() time.Time {
				return time.Time{}
			},
			"isZeroTime": func(t time.Time) bool {
				return t.IsZero()
			},
			"formatDuration": func(d time.Duration) string {
				return d.String()
			},
		}
	})
}

// allSessionsWrapper allows rendering an HTML page that lists all charts
type allSessionsWrapper struct {
	babyapi.ResourceList[*SessionResource]
}

func (as allSessionsWrapper) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (as allSessionsWrapper) HTML(_ http.ResponseWriter, r *http.Request) string {
	slices.SortFunc(as.Items, func(a, b *SessionResource) int {
		return b.Session.StartTime.Compare(a.Session.StartTime)
	})
	return listSessions.Render(r, as.Items)
}
