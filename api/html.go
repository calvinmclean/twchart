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
    <script src="https://unpkg.com/htmx.org@1.9.8"></script>
</head>
<body class="uk-background-muted uk-padding">
    <div class="uk-container uk-container-small" id="sessions-container">
	    <ul class="uk-breadcrumb uk-margin-small-top">
		    <li><a href="/sessions">Sessions</a></li>
		    <li><span></span></li>
		</ul>

		<div class="uk-flex uk-flex-between uk-flex-middle">
            <h1 class="uk-heading-line"><span>Sessions</span></h1>
        </div>

	<!-- Type Filter -->
        <div class="uk-margin uk-flex uk-flex-middle">
            <select id="type-filter" name="type" class="uk-select uk-form-width-medium"
                    hx-get="/sessions" hx-target="#sessions-container" hx-swap="outerHTML" hx-push-url="true"
                    hx-headers='{"Accept": "text/html"}'>
                <option value="">All Types</option>
                {{ $types := getUniqueTypes .Sessions }}
                {{ range $types }}
                <option value="{{ . }}" {{ if eq . $.Pagination.Type }}selected{{ end }}>{{ . }}</option>
                {{ end }}
            </select>
        </div>

        {{ if .Sessions }}
        <ul id="sessions-list" class="uk-list uk-list-divider uk-margin">
            {{ range .Sessions }}
            <li class="uk-flex uk-flex-between uk-flex-middle session-item">
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

        {{ template "pagination" . }}

        {{ else }}
        <p class="uk-text-center uk-text-muted">No sessions found.</p>
        {{ end }}

    </div>

</body>
</html>
{{ end }}`

	pagination         = html.Template("pagination")
	paginationTemplate = `{{ define "pagination" }}
{{ if gt .Pagination.TotalPages 1 }}
<div class="uk-flex uk-flex-center uk-margin">
    <ul class="uk-pagination">
        <!-- Previous -->
        {{ if .Pagination.HasPrev }}
        <li>
            <a hx-get="/sessions?page={{ .Pagination.PrevPage }}&per_page={{ .Pagination.PerPage }}{{ if .Pagination.Type }}&type={{ .Pagination.Type }}{{ end }}"
               hx-target="#sessions-container"
               hx-swap="outerHTML"
               hx-push-url="true"
               hx-headers='{"Accept": "text/html"}'>
                <span uk-pagination-previous></span>
            </a>
        </li>
        {{ else }}
        <li class="uk-disabled"><span uk-pagination-previous></span></li>
        {{ end }}

        <!-- Page Numbers -->
        {{ $pages := getPageRange .Pagination.Page .Pagination.TotalPages }}
        {{ range $pages }}
            {{ if eq . -1 }}
            <li class="uk-disabled"><span>...</span></li>
            {{ else if eq . $.Pagination.Page }}
            <li class="uk-active"><span>{{ . }}</span></li>
            {{ else }}
            <li>
                <a hx-get="/sessions?page={{ . }}&per_page={{ $.Pagination.PerPage }}{{ if $.Pagination.Type }}&type={{ $.Pagination.Type }}{{ end }}"
                   hx-target="#sessions-container"
                   hx-swap="outerHTML"
                   hx-push-url="true"
                   hx-headers='{"Accept": "text/html"}'>{{ . }}</a>
            </li>
            {{ end }}
        {{ end }}

        <!-- Next -->
        {{ if .Pagination.HasNext }}
        <li>
            <a hx-get="/sessions?page={{ .Pagination.NextPage }}&per_page={{ .Pagination.PerPage }}{{ if .Pagination.Type }}&type={{ .Pagination.Type }}{{ end }}"
               hx-target="#sessions-container"
               hx-swap="outerHTML"
               hx-push-url="true"
               hx-headers='{"Accept": "text/html"}'>
                <span uk-pagination-next></span>
            </a>
        </li>
        {{ else }}
        <li class="uk-disabled"><span uk-pagination-next></span></li>
        {{ end }}
    </ul>
</div>

<p class="uk-text-center uk-text-meta">
    Page {{ .Pagination.Page }} of {{ .Pagination.TotalPages }} 
    ({{ .Pagination.StartItem }} - {{ .Pagination.EndItem }} of {{ .Pagination.Total }} sessions)
</p>
{{ end }}
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
		string(pagination):    paginationTemplate,
	})

	html.SetFuncs(func(r *http.Request) map[string]any {
		return map[string]any{
			"getUniqueTypes": getUniqueTypes,
			"getPageRange":   getPageRange,
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

// getPageRange generates a smart list of page numbers for pagination display
// Returns slice of page numbers, with -1 indicating ellipsis
func getPageRange(currentPage, totalPages int64) []int64 {
	if totalPages <= 5 {
		// Show all pages if 5 or fewer
		pages := make([]int64, totalPages)
		for i := int64(0); i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages
	}

	var pages []int64

	// Always show first page
	pages = append(pages, 1)

	if currentPage <= 3 {
		// Near the beginning: show 1, 2, 3, 4, ..., last
		pages = append(pages, 2, 3, 4)
		pages = append(pages, -1) // ellipsis
	} else if currentPage >= totalPages-2 {
		// Near the end: show 1, ..., last-3, last-2, last-1, last
		pages = append(pages, -1) // ellipsis
		pages = append(pages, totalPages-3, totalPages-2, totalPages-1)
	} else {
		// Middle: show 1, ..., current-1, current, current+1, ..., last
		pages = append(pages, -1) // ellipsis
		pages = append(pages, currentPage-1, currentPage, currentPage+1)
		pages = append(pages, -1) // ellipsis
	}

	// Always show last page
	pages = append(pages, totalPages)

	return pages
}

// SessionsListData holds the data for rendering the sessions list template
type SessionsListData struct {
	Sessions   []*SessionResource
	Pagination PaginationParams
}

// allSessionsWrapper allows rendering an HTML page that lists all charts
type allSessionsWrapper struct {
	babyapi.ResourceList[*SessionResource]
	sessions []*SessionResource
}

func (as allSessionsWrapper) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (as allSessionsWrapper) HTML(w http.ResponseWriter, r *http.Request) string {
	// Get pagination params from context
	params := getPaginationParams(r.Context())

	// Sort sessions by start time
	slices.SortFunc(as.sessions, func(a, b *SessionResource) int {
		return b.Session.StartTime.Compare(a.Session.StartTime)
	})

	// Get total count from storage adapter
	api := getAPIFromContext(r.Context())
	if api != nil && api.storageAdapter.Client != nil {
		total, err := api.storageAdapter.GetTotalCount(r.Context(), params.Type)
		if err == nil {
			params.Total = total
		}
	} else {
		// For non-SQL storage, count the sessions
		params.Total = int64(len(as.sessions))

		// Apply pagination slicing for non-SQL storage
		start := params.Offset()
		end := start + params.PerPage
		if end > int64(len(as.sessions)) {
			end = int64(len(as.sessions))
		}
		if start < int64(len(as.sessions)) {
			as.sessions = as.sessions[start:end]
		} else {
			as.sessions = nil
		}
	}

	data := SessionsListData{
		Sessions:   as.sessions,
		Pagination: params,
	}

	return listSessions.Render(r, data)
}
