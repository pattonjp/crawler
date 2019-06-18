package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

var mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

func (s *Server) loadTemplates() error {
	dir := s.templateDir
	if s.templates == nil {
		s.templates = make(map[string]*template.Template)
	}

	// dir := "templates/"

	layouts, err := filepath.Glob(dir + "/layouts/*.gohtml")
	if err != nil {
		return err
	}

	includes, err := filepath.Glob(dir + "/*.gohtml")
	if err != nil {
		return err
	}
	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		return err
	}
	for _, include := range includes {
		files := append(layouts, include)
		name := strings.Replace(filepath.Base(include), ".gohtml", "", 1)
		s.templates[name], err = mainTemplate.Clone()
		if err != nil {
			return err
		}
		s.templates[name] = template.Must(s.templates[name].ParseFiles(files...))
	}

	return nil
}

func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
	if s.debug {
		_ = s.loadTemplates()
	}
	tmpl, ok := s.templates[name]
	if !ok {
		s.handleError(w)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	if err := tmpl.Execute(w, data); err != nil {
		fmt.Println(err)
		s.handleError(w)
		return
	}

}
func (s *Server) handleError(w http.ResponseWriter) {
	http.Error(w, "unhandled error", http.StatusInternalServerError)
}
