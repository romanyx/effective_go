package main

import (
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
)

// GzipOn enables gzip compression for Server
func GzipOn(s *Server) {
	s.gzip = true
}

// ReadTimeout sets Server response read timeout
func ReadTimeout(timeout time.Duration) func(*Server) {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// GET sets Server GET Handle for specified path
func GET(path string, handle httprouter.Handle) func(*Server) {
	return func(s *Server) {
		s.router.GET(path, handle)
	}
}

// Server is a http server
type Server struct {
	server http.Server
	router *httprouter.Router
	gzip   bool
}

// ListenAndServe calls Server http.Server ListenAndServe
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// NewServer returns initialized Server
func NewServer(addr string, options ...func(*Server)) *Server {
	s := newServer(addr)

	for _, option := range options {
		option(&s)
	}

	if s.gzip {
		s.server.Handler = gziphandler.GzipHandler(s.router)
	}

	return &s
}

func newServer(addr string) Server {
	router := httprouter.New()

	return Server{
		server: http.Server{
			Addr:    addr,
			Handler: router,
		},
		router: router,
	}
}
