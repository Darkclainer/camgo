package querier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/darkclainer/camgo/pkg/parser"
)

func errorRequestf(t *testing.T, w http.ResponseWriter, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	t.Error(str)
	http.Error(w, str, http.StatusInternalServerError)
}

func newTestRemote( // nolint:gocritic // test
	t *testing.T,
	queryFn, suggestionFn, lemmaFn map[string]http.HandlerFunc,
) (
	*Remote,
	func(),
) {
	mux := http.NewServeMux()
	mux.HandleFunc(suggestionPath, func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()
		query := queryValues.Get("q")
		if query == "" {
			errorRequestf(t, w, "request with empty q query")
			return
		}
		fn, ok := suggestionFn[query]
		if !ok {
			errorRequestf(t, w, "handler for suggestion '%s' not found", query)
			return
		}
		fn(w, r)
	})
	mux.HandleFunc(searchPath, func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()
		query := queryValues.Get("q")
		if query == "" {
			errorRequestf(t, w, "request with empty q query")
			return
		}
		fn, ok := queryFn[query]
		if !ok {
			errorRequestf(t, w, "handler for query '%s' not found", query)
			return
		}
		fn(w, r)
	})
	mux.HandleFunc(lemmaPath, func(w http.ResponseWriter, r *http.Request) {
		rawPath := r.URL.Path
		dir, lemmaID := path.Split(rawPath)
		if dir != lemmaPath {
			errorRequestf(t, w, "lemma invalid lemma request '%s'", rawPath)
			return
		}
		fn, ok := lemmaFn[lemmaID]
		if !ok {
			errorRequestf(t, w, "handler for lemma '%s' not found", lemmaID)
			return
		}
		fn(w, r)
	})
	server := httptest.NewServer(mux)
	client := server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	querier := NewRemote(client, &JSONParser{}, &Config{
		Host:     server.Listener.Addr().String(),
		Protocol: "http",
	})
	return querier, func() {
		server.Close()
		_ = querier.Close(context.TODO())
	}
}

func redirectSuggestions(w http.ResponseWriter, r *http.Request, query string, status int) {
	u := *r.URL
	u.Path = suggestionPath
	values := url.Values{}
	values.Add("q", query)
	u.RawQuery = values.Encode()

	http.Redirect(w, r, u.String(), status)
}

func TestRemoteSearch(t *testing.T) { // nolint:funlen // test
	testCases := map[string]struct {
		query        string
		lemmaID      string
		suggestions  []string
		queryFn      map[string]http.HandlerFunc
		suggestionFn map[string]http.HandlerFunc
		err          error
	}{
		"return lemmaID": {
			query:   "hello",
			lemmaID: "imhere",
			queryFn: map[string]http.HandlerFunc{
				"hello": func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, path.Join(lemmaPath, "imhere"), http.StatusFound)
				},
			},
		},
		"return empty lemmaID": {
			query:   "hello",
			lemmaID: "",
			queryFn: map[string]http.HandlerFunc{
				"hello": func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, lemmaPath, http.StatusFound)
				},
			},
			err: ErrEmptyLemmaID,
		},
		"return lemmaID wrong status": {
			query: "wrong status",
			queryFn: map[string]http.HandlerFunc{
				"wrong status": func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, path.Join(lemmaPath, "imhere"), http.StatusOK)
				},
			},
			err: errors.New("error"),
		},
		"return suggestions": {
			query:       "helo",
			suggestions: []string{"hello", "hell"},
			queryFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					redirectSuggestions(w, r, "helo", http.StatusFound)
				},
			},
			suggestionFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					suggestions := []string{"hello", "hell"}
					err := json.NewEncoder(w).Encode(suggestions)
					assert.NoError(t, err)
				},
			},
		},
		"return empty suggestions": {
			query: "helo",
			queryFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					redirectSuggestions(w, r, "helo", http.StatusFound)
				},
			},
			suggestionFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					suggestions := []string{}
					err := json.NewEncoder(w).Encode(suggestions)
					assert.NoError(t, err)
				},
			},
			err: errors.New("error"),
		},
		"return malformed suggestions": {
			query: "helo",
			queryFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					redirectSuggestions(w, r, "helo", http.StatusFound)
				},
			},
			suggestionFn: map[string]http.HandlerFunc{
				"helo": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("{,}")) // json error decoding
				},
			},
			err: errors.New("error"),
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			querier, clean := newTestRemote(t, tc.queryFn, tc.suggestionFn, nil)
			defer clean()

			lemmaID, suggestions, err := querier.Search(context.TODO(), tc.query)
			switch {
			case tc.err != nil:
				assert.Error(t, err)
				return
			case len(tc.suggestions) > 0:
				if errors.Is(err, ErrSuggestions) {
					assert.Equal(t, tc.suggestions, suggestions)
					return
				}
				t.Errorf("suggestion must be returned, got: %v", err)
				return
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.lemmaID, lemmaID)
		})
	}
}
func TestRemoteGetLemma(t *testing.T) {
	testLemmas := []*parser.Lemma{
		{
			Lemma:     "test lemma",
			GuideWord: "hello",
		},
	}
	testCases := map[string]struct {
		lemmaID string
		lemmas  []*parser.Lemma
		lemmaFn map[string]http.HandlerFunc
		err     error
	}{
		"return lemmas": {
			lemmaID: "hello",
			lemmas:  testLemmas,
			lemmaFn: map[string]http.HandlerFunc{
				"hello": func(w http.ResponseWriter, r *http.Request) {
					err := json.NewEncoder(w).Encode(testLemmas)
					assert.NoError(t, err)
				},
			},
		},
		"return malformed lemmas": {
			lemmaID: "hello",
			lemmaFn: map[string]http.HandlerFunc{
				"hello": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("{,}")) // json error decoding
				},
			},
			err: errors.New("error"),
		},
		"return wrong status code": {
			lemmaID: "hello",
			lemmaFn: map[string]http.HandlerFunc{
				"hello": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				},
			},
			err: errors.New("error"),
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			querier, clean := newTestRemote(t, nil, nil, tc.lemmaFn)
			defer clean()

			lemmas, err := querier.GetLemma(context.TODO(), tc.lemmaID)
			if tc.err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.lemmas, lemmas)
		})
	}
}
