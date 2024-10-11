// Package handlers contains the HTTP request handlers for the server.
// Each handler function is responsible for processing a specific type of HTTP request and generating a response.
package handlers

import (
	"html/template"
	"net/http"
)

var getAllMetricsTemplate = `
<table>
<thead>
	
</thead>
<tbody>
{{range $metric := .}}
     <tr>
	 	<td>{{.Name}}</td>
		<td>{{.Counter}}</td>
		<td>{{.Gauge}}</td>
	 </tr>
{{end}}
</tbody>
</table>
`

// GetAll returns all metrics in html.
func (h *MetricHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	metrics, err := h.repository.FindAll(r.Context())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")

	t := template.Must(template.New("tmpl").Parse(getAllMetricsTemplate))

	t.Execute(rw, metrics)
}
