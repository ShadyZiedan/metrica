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
		<td>{{.GetCounter}}</td>
		<td>{{.GetGauge}}</td>
	 </tr>
{{end}}
</tbody>
</table>
`

func (h *MetricHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	metrics, err := h.repository.FindAll()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")

	t := template.Must(template.New("tmpl").Parse(getAllMetricsTemplate))

	t.Execute(rw, metrics)
}
