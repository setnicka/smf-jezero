package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Execute template given by its name and with given data with all the error handling.
func (s *Server) executeTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	slog := slog.With("template", templateName)

	slog.Debug("executing template")
	template, err := s.getTemplates()
	if err != nil || template == nil {
		slog.Error("error getting templates", "err", err)
		fmt.Fprintf(w, "Error getting templates: %v", err)
		return
	}
	err = template.ExecuteTemplate(w, templateName, data)
	if err != nil {
		slog.Error("error executing template", "err", err)
		fmt.Fprintf(w, "Error executing template. %v", err)
	}
}

// Scan directory with templates and if there is some changed file reload all templates,
// then return these loaded templates.
func (s *Server) getTemplates() (*template.Template, error) {
	globPath := path.Join(s.cfg.TemplateDir, "*.tmpl")
	templateFiles, err := filepath.Glob(globPath)
	if err != nil {
		return nil, err
	}
	changed := false
	for _, file := range templateFiles {
		if fileChanged(file) {
			slog.Debug("found (new/changed) template", "file", file)
			changed = true
		}
	}

	if changed {
		slog.Debug("parsing all template files because of new/changed template files")
		s.templates, err = template.New("").Funcs(templateFuncs).ParseGlob(globPath)
		if err != nil {
			return nil, err
		}
	}
	return s.templates, nil
}

// Hashes are not computed on every request - hashes are remembered and they are
// recomputed only when mod time of file changes
type fileHashInfo struct {
	modTime time.Time
	hash    string
}

var fileModMap = make(map[string]fileHashInfo)

func fileChanged(path string) bool {
	stats, err := os.Stat(path)
	if err != nil {
		return true // missing file is although change
	}

	record, exists := fileModMap[path]
	if !exists || record.modTime != stats.ModTime() {
		newField := fileHashInfo{stats.ModTime(), ""} // no need to compute hash
		fileModMap[path] = newField
		return true
	}
	return false
}
