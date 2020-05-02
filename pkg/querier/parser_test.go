package querier

import (
	"encoding/json"
	"io"

	"github.com/darkclainer/camgo/pkg/parser"
)

// JSONParser parses lemma or suggestions from JSON format. Use it for testing
type JSONParser struct{}

func (p *JSONParser) ParseLemma(page io.Reader) ([]*parser.Lemma, error) {
	var lemmas []*parser.Lemma
	if err := json.NewDecoder(page).Decode(&lemmas); err != nil {
		return nil, err
	}
	return lemmas, nil
}

func (p *JSONParser) ParseSuggestion(page io.Reader) ([]string, error) {
	var suggestions []string
	if err := json.NewDecoder(page).Decode(&suggestions); err != nil {
		return nil, err
	}
	return suggestions, nil
}
