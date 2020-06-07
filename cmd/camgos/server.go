package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/darkclainer/camgo/pkg/querier"
)

type Server struct {
	http.Server
	mux    http.ServeMux
	conf   *Config
	logger *zap.Logger
	q      querier.Querier
}

func New(logger *zap.Logger, conf *Config) (*Server, error) {
	s := Server{
		conf:   conf,
		logger: logger,
	}

	var q querier.Querier
	q = querier.NewRemote(nil, nil, &conf.Remote)
	if conf.Cached.Path != "" || conf.Cached.InMemory != false {
		cached, err := querier.NewCached(q, &conf.Cached)
		if err != nil {
			return nil, err
		}
		q = cached
	}
	s.q = q

	s.mux.HandleFunc("/query", s.middleLogging(s.handleQuery()))
	s.mux.HandleFunc("/lemma", s.middleLogging(s.handleQuery()))
	s.Addr = conf.Host
	s.Server.Handler = &s.mux
	return &s, nil
}

func (s *Server) Close(ctx context.Context) error {
	var reasons []string
	if serverErr := s.Server.Shutdown(ctx); serverErr != nil {
		reasons = append(reasons, "server shutdown failed: "+serverErr.Error())
	}
	if querierErr := s.q.Close(ctx); querierErr != nil {
		reasons = append(reasons, "querier close failed: "+querierErr.Error())
	}
	if len(reasons) > 0 {
		return fmt.Errorf("close failed because: %s", strings.Join(reasons, " AND "))
	}
	return nil

}

func (s *Server) respondJSON(w http.ResponseWriter, vPtr interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(vPtr); err != nil {
		s.logger.Error("encodig failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"encoding error"}`))
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(buffer.Bytes())
}

func (s *Server) middleLogging(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("request",
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("client", r.RemoteAddr),
			zap.String("method", r.Method),
		)
		handler(w, r)
	}
}
