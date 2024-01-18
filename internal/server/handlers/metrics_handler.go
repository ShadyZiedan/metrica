package handlers

import (
	"html/template"
	"net/http"

	"github.com/shadyziedan/metrica/internal/repositories"
)

const metricsTemplate = `
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

func MetricsHandler(repo repositories.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metrics, err := repo.FindAll()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		t := template.Must(template.New("tmpl").Parse(metricsTemplate))

		t.Execute(rw, metrics)
	}
}
