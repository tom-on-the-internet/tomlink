package main

// Registers route handlers.
func (s *server) routes() {
	s.router.Get("/", s.handleCreateRedirect())
	s.router.Get("/redirects/{accessCode:[a-f0-9]+}", s.handleViewRedirect())
	s.router.Get("/failure", s.handleFailure())
	s.router.Get("/deleted", s.handleDeleted())
	s.router.Get("/identify", s.handleIdentify())
	s.router.Get("/static/*", s.handleStatic())
	s.router.Get("/{link:[a-z0-9-]{3,}}", s.handleRedirect())
	s.router.Post("/", s.handlePostRoot())
	s.router.Post("/redirects/{accessCode:[a-f0-9]+}/delete", s.handleDelete())
	s.router.Get("/*", s.handleNotFound())
	s.router.Options("/*", s.handleCORS())
}
