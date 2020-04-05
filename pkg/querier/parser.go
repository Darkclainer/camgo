package querier

import (
	"io"

	"github.com/darkclainer/camgo/pkg/parser"
)

type Parser interface {
	ParseLemma(page io.Reader) ([]*parser.Lemma, error)
	ParseSuggestion(page io.Reader) ([]string, error)
}

type HTMLParser struct{}

func (p *HTMLParser) ParseLemma(page io.Reader) ([]*parser.Lemma, error) {
	return parser.ParseLemmaHTML(page)
}
func (p *HTMLParser) ParseSuggestion(page io.Reader) ([]string, error) {
	return parser.ParseSuggestionHTML(page)
}
