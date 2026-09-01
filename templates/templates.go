package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/sandbanks/beads-everywhere/pkg/models"
)

//go:embed layouts/*.html pages/*.html components/*.html
var FS embed.FS

var (
	funcMap = template.FuncMap{
		"priorityBadgeClass": func(iss models.Issue) string {
			return iss.PriorityBadgeClass()
		},
		"priorityLabel": func(iss models.Issue) string {
			return iss.PriorityLabel()
		},
		"hasPrefix": strings.HasPrefix,
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
	}

	tmpl = template.Must(template.New("").Funcs(funcMap).ParseFS(FS,
		"layouts/*.html",
		"pages/*.html",
		"components/*.html",
	))
)

// Render executes a named template with data into writer w.
func Render(w io.Writer, name string, data any) error {
	return tmpl.ExecuteTemplate(w, name, data)
}

// PageData represents the view model for the full index page.
type PageData struct {
	Title         string
	Projects      []models.Project
	SelectedRepo  string
	CurrentFilter string
	Search        string
	Issues        []models.Issue
}

// TabsData represents data for the filter tabs partial.
type TabsData struct {
	SelectedRepo  string
	CurrentFilter string
	OOB           bool
}

// TabsWithIssuesData represents data for OOB swaps containing tabs + issue list.
type TabsWithIssuesData struct {
	SelectedRepo  string
	CurrentFilter string
	OOB           bool
	Issues        []models.Issue
}
