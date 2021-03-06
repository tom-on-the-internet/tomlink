package main

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	urlpkg "net/url"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var templateFiles embed.FS

func (s *server) handleStatic() http.HandlerFunc {
	staticFS := http.FS(staticFiles)
	fs := http.FileServer(staticFS)

	return func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
}

func (s *server) handleCreateRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.RenderTemplate(w, http.StatusOK, newViewData("create"))
	}
}

func (s *server) handleViewRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessCode := chi.URLParam(r, "accessCode")

		redirect, err := s.repository.getRedirectByAccessCode(accessCode)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				s.NotFound(w, r)
			}

			s.Failure(w, r)

			return
		}

		vd := newViewData("view-redirect")
		vd.Redirect = redirect
		vd.CurrentURL = template.URL(vd.Host + "/redirects/" + vd.Redirect.AccessCode)
		vd.LinkURL = template.URL(vd.Host + "/" + vd.Redirect.Link)

		s.RenderTemplate(w, http.StatusOK, vd)
	}
}

func (s *server) handleRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		link := chi.URLParam(r, "link")

		redirect, err := s.repository.getRedirectByLink(link)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				s.NotFound(w, r)
			}

			s.Failure(w, r)

			return
		}

		ipAdress := strings.Split(r.RemoteAddr, ":")[0]

		visit, err := createVisitFromIPAddress(ipAdress)
		if err != nil {
			http.Redirect(w, r, redirect.URL, http.StatusFound)
		} else {
			s.repository.createVisit(redirect, visit)
			http.Redirect(w, r, redirect.URL, http.StatusFound)
		}
	}
}

func (s *server) handleFailure() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Failure(w, r)
	}
}

func (s *server) handlePostRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)

			s.Failure(w, r)

			return
		}

		link := r.FormValue("link")
		url := r.FormValue("url")

		_, notExistsErr := s.repository.getRedirectByLink(link)

		isValidLink, _ := regexp.MatchString("[a-z0-9-]{3,}", link)

		// check possible reasons why we should not allow this link
		if isReservedWord(link) || notExistsErr == nil || !isValidLink {
			vd := newViewData("invalid-link")
			rd := Redirect{Link: link, URL: url}
			vd.Redirect = rd

			s.RenderTemplate(w, http.StatusBadRequest, vd)

			return
		}

		_, urlIsInvalidError := urlpkg.ParseRequestURI(url)
		if !urlIsValid(url) || urlIsInvalidError != nil {
			vd := newViewData("invalid-url")
			rd := Redirect{Link: link, URL: url}
			vd.Redirect = rd

			s.RenderTemplate(w, http.StatusBadRequest, vd)

			return
		}

		redirect, err := s.repository.createRedirect(link, url)
		if err != nil {
			http.Redirect(w, r, "/failure", http.StatusSeeOther)

			return
		}

		http.Redirect(w, r, "/redirects/"+redirect.AccessCode, http.StatusSeeOther)
	}
}

func (s *server) handleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessCode := chi.URLParam(r, "accessCode")

		err := s.repository.deleteRedirectByAccessCode(accessCode)
		if err != nil {
			http.Redirect(w, r, "/failure", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/deleted", http.StatusSeeOther)
	}
}

func (s *server) handleDeleted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.RenderTemplate(w, http.StatusOK, newViewData("deleted"))
	}
}

func (s *server) handleNotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.NotFound(w, r)
	}
}

// Checks if the link is one that would either break the application, or make Tom unhappy.
func isReservedWord(str string) bool {
	reservedWords := []string{"tomlink", "deleted", "failure", "static", "redirects", "redirect", "delete", "deleted", "404", "not-found"}

	for _, val := range reservedWords {
		if val == str {
			return true
		}
	}

	return false
}

func urlIsValid(str string) bool {
	resp, err := http.Get(str)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return false
	}

	return true
}
