package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/coreos/go-log/log"
)

// Execute template given by its name and with given data with all the error handling.
func executeTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	log.Debugf("Executing template '%s'", templateName)
	template, err := server.getTemplates()
	if err != nil || template == nil {
		log.Errorf("Error getting templates: %v ", err)
		fmt.Fprintf(w, "Error getting templates: %v", err)
		return
	}
	err = template.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Errorf("Error executing template '%s': %v", templateName, err)
		fmt.Fprintf(w, "Error executing template. %v", err)
	}
}

// Scan directory with templates and if there is some changed file reload all templates,
// then return these loaded templates.
func (s *Server) getTemplates() (*template.Template, error) {
	globPath := path.Join(TEMPLATE_DIR, "*.tmpl")
	templateFiles, err := filepath.Glob(globPath)
	if err != nil {
		return nil, err
	}
	changed := false
	for _, file := range templateFiles {
		if fileChanged(file) {
			log.Debugf("Found (new/changed) template file '%s'", file)
			changed = true
		}
	}

	if changed {
		log.Debug("Parsing all template files because of new/changed template files")
		s.templates, err = template.ParseGlob(globPath)
		if err != nil {
			return nil, err
		}
	}
	return s.templates, nil
}

// Hashes are not computed on every request - hashes are remebered and they are
// recomputed only when mod time of file changes
type fileHashInfo struct {
	modTime time.Time
	hash    string
}

var fileModMap map[string]fileHashInfo = make(map[string]fileHashInfo)

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
