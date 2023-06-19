package server

import (
	"html/template"
)

var (
	templateFuncs = template.FuncMap{
		"mult": func(a, b int) int {
			return a * b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
)
