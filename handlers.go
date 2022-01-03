package main

import (
	"embed"
	"encoding/json"
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

		ipAddress := strings.Split(r.RemoteAddr, ":")[0]

		visit, err := createVisitFromIPAddress(ipAddress)
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

func (s *server) handleIdentify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipAddress := strings.Split(r.RemoteAddr, ":")[0]

		visit := Visit{IPAddress: ipAddress}

		res, err := http.Get("http://ip-api.com/json/" + ipAddress)
		if err != nil {
			log.Println(err)
			s.handleFailure()
		}

		defer res.Body.Close()

		err = json.NewDecoder(res.Body).Decode(&visit)
		if err != nil {
			log.Println(err)
			s.handleFailure()
		}

		origins, ok := r.Header["Origin"]
		if !ok || len(origins) == 0 {
			log.Println(err)
			s.handleFailure()
		}

		origin := origins[0]

		if origin != "https://tomontheinternet.com" && origin != "https://www.tomontheinternet.com" && origin != "http://127.0.0.1:8080/" {
			log.Println(err)
			s.handleFailure()
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", origin)
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		_ = json.NewEncoder(w).Encode(visit)
	}
}

func (s *server) handleCORS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "GET")

		w.Write([]byte("{\"message\": \"success\"}"))
	}
}

// Checks if the link is one that would either break the application, or make Tom unhappy.
func isReservedWord(str string) bool {
	reservedWords := []string{"tomlink", "deleted", "failure", "static", "redirects", "redirect", "delete", "deleted", "404", "not-found", "identify"}

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
