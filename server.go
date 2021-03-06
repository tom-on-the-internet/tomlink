package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type server struct {
	router     *chi.Mux
	port       string
	repository repository
	template   *template.Template
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) Close() {
	s.repository.close()
}

func (s *server) NotFound(w http.ResponseWriter, r *http.Request) {
	vd := newViewData("not-found")

	s.RenderTemplate(w, http.StatusNotFound, vd)
}

func (s *server) Failure(w http.ResponseWriter, r *http.Request) {
	vd := newViewData("failure")

	s.RenderTemplate(w, http.StatusInternalServerError, vd)
}

func (s *server) RenderTemplate(w http.ResponseWriter, statusCode int, vd ViewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	err := s.template.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Println(err)
	}
}

func newServer() *server {
	s := &server{
		router: chi.NewRouter(), port: ":3000", repository: newRepository(),
		template: template.Must(template.ParseFS(templateFiles, "templates/*")),
	}
	s.routes()

	return s
}
