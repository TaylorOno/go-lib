package metrics

import (
	"html/template"
	"net/http"
	"sort"
	"strings"
)

const metricTpl = `# Service metrics
| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
{{- range .Definitions }}
| {{.Name}} | {{.Definition.Description}} | {{.Definition.Kind}} | {{.Definition.Labels | splitter }} |
{{- end }}
`

var (
	splitter = func(s []string) string { return strings.Join(s, ", ") }
)

type metricDefinition struct {
	Name       string
	Definition MetricDefinition
}

// MetricDocs produces a file using a custom reporter
func MetricDocs(metricsReporter Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := metricsReporter.GetMetricsDefinition()
		values := make([]metricDefinition, len(metrics))
		ndx := 0
		for k, v := range metrics {
			values[ndx] = metricDefinition{Name: k, Definition: v}
			ndx = ndx + 1
		}

		definition := struct{ Definitions []metricDefinition }{Definitions: values}

		sort.Slice(definition.Definitions[:], func(i, j int) bool {
			return definition.Definitions[i].Name < definition.Definitions[j].Name
		})

		report, err := template.New("metrics").Funcs(template.FuncMap{"splitter": splitter}).Parse(metricTpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		report.Execute(w, definition)
	}
}
