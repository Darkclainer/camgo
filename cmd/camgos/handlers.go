package main

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/darkclainer/camgo/pkg/parser"
	"github.com/darkclainer/camgo/pkg/querier"
)

type ResponseStatus int

const (
	ResponseOK ResponseStatus = iota
	ResponseSuggestions
	ResponseBadRequest
	ResponseError
)

type ResponseSearch struct {
	LemmaID     string          `json:"lemma_id,omitempty"`
	Lemmas      []*parser.Lemma `json:"lemma,omitempty"`
	Suggestions []string        `json:"suggestions,omitempty"`
	Status      ResponseStatus  `json:"status"`
}

func (s *Server) handleQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		query, ok := r.URL.Query()["q"]
		if !ok || len(query) < 1 {
			s.respondJSON(w, &ResponseSearch{
				Status: ResponseBadRequest,
			}, http.StatusBadRequest)
			return
		}
		lemmaID, suggestions, err := s.q.Search(r.Context(), query[0])
		response := ResponseSearch{
			LemmaID:     lemmaID,
			Suggestions: suggestions,
		}
		if err != nil {
			response.Status = ResponseError
			if errors.Is(err, querier.ErrSuggestions) {
				response.Status = ResponseSuggestions
			} else {
				s.logger.Error("Querier search returned error",
					zap.Error(err),
					zap.String("query", query[0]),
				)
			}
		}
		if lemmaID != "" {
			lemmas, err := s.q.GetLemma(r.Context(), lemmaID)
			response.Lemmas = lemmas
			if err != nil {
				response.Status = ResponseError
				s.logger.Error("Querier GetLemma returned error",
					zap.Error(err),
					zap.String("lemma_id", lemmaID),
				)
			}
		}
		s.respondJSON(w, &response, http.StatusOK)
	}
}
