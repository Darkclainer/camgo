package camgo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultHost    = "dictionary.cambridge.org"
	lemmaPath      = "/dictionary/english/"
	suggestionPath = "/spellcheck/english/"
	searchPath     = "/search/english/direct/"
)

type ErrLemmaNotFound []string

func (e ErrLemmaNotFound) Error() string {
	return "lemma not found"
}

func (e ErrLemmaNotFound) Suggestions() []string {
	return e
}

type QuerierConfig struct {
	ExtraHeader map[string]string
	Timeout     time.Duration
	Host        string
}

type Querier struct {
	client *http.Client
	config *QuerierConfig
}

func NewQuerier(client *http.Client, config *QuerierConfig) *Querier {
	if client == nil {
		client = getDefaultQuerierClient()
	}
	if config.Host == "" {
		config.Host = defaultHost
	}
	return &Querier{
		client: client,
		config: config,
	}
}

// getDefaultQuerierClient returns default client for querier that ignores redirect
func getDefaultQuerierClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (q *Querier) GetLemma(ctx context.Context, lemmaID string) ([]*Lemma, error) {
	response, err := q.get(ctx, q.newLemmaURL(lemmaID), http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to get lemma: %w", err)
	}
	defer response.Body.Close()
	lemmas, err := ParseLemmaHTML(response.Body)
	if err != nil {
		return nil, err
	}
	return lemmas, nil
}

// Search returns lemmaID if found something
// Also it can return ErrLemmaNotFound error if there is some suggestions
func (q *Querier) Search(ctx context.Context, query string) (string, error) {
	redirect, err := q.getSearch(ctx, q.newSearchURL(query))
	if err != nil {
		return "", fmt.Errorf("can not perform search: %w", err)
	}
	switch {
	case strings.HasPrefix(redirect.Path, lemmaPath):
		return path.Base(redirect.Path), nil
	case strings.HasPrefix(redirect.Path, suggestionPath):
		suggestions, err := q.getSuggestions(ctx, redirect.String())
		if err != nil {
			return "", fmt.Errorf("can not get suggestions: %w", err)
		}
		return "", ErrLemmaNotFound(suggestions)
	default:
		return "", fmt.Errorf("uknown redirect: %s", redirect)
	}
}

func (q *Querier) getSearch(ctx context.Context, urlSearch string) (*url.URL, error) {
	response, err := q.get(ctx, urlSearch, http.StatusFound)
	if err != nil {
		return nil, fmt.Errorf("failed to perform get: %w", err)
	}
	defer response.Body.Close()
	redirect, err := url.Parse(response.Header.Get("Location"))
	if err != nil {
		return nil, fmt.Errorf("can not parse redirect url: %w", err)
	}
	return redirect, nil
}

func (q *Querier) getSuggestions(ctx context.Context, urlSuggestions string) ([]string, error) {
	response, err := q.get(ctx, urlSuggestions, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to perform get: %w", err)
	}
	defer response.Body.Close()
	suggestions, err := ParseSuggestionHTML(response.Body)
	if err != nil {
		return nil, err
	}
	return suggestions, nil
}

func (q *Querier) get(ctx context.Context, urlGet string, expectedStatus int) (*http.Response, error) {
	request, err := q.newRequest(ctx, urlGet)
	if err != nil {
		return nil, fmt.Errorf("can not assemble request: %w", err)
	}
	response, err := q.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	if response.StatusCode != expectedStatus {
		response.Body.Close()
		return nil, fmt.Errorf("unexpected response code: %d", response.StatusCode)
	}
	return response, err
}

func (q *Querier) newSearchURL(query string) string {
	searchURL := q.newURL()
	searchURL.Path = searchPath

	v := url.Values{}
	v.Set("q", query)
	v.Set("datasetsearch", "english")
	searchURL.RawQuery = v.Encode()

	return searchURL.String()
}

func (q *Querier) newLemmaURL(lemmaID string) string {
	lemmaURL := q.newURL()
	lemmaURL.Path = path.Join(lemmaPath, lemmaID)
	return lemmaURL.String()
}

func (q *Querier) newURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   q.config.Host,
	}
}

func (q *Querier) newRequest(ctx context.Context, urlRequest string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlRequest, nil)
	if err != nil {
		return nil, fmt.Errorf("can not form request: %w", err)
	}
	for key, value := range q.config.ExtraHeader {
		req.Header.Add(key, value)
	}
	return req, nil
}
