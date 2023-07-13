package server

import (
	"fmt"
	"html/template"
)

var (
	templateFuncs = template.FuncMap{
		"mult": func(a, b int) int {
			return a * b
		},
		"percentage": func(a, max int) string {
			return fmt.Sprintf("%.2f%%", float64(a)/float64(max)*100)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
)
