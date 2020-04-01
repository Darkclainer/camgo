package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

var suggestionListMatcher = cascadia.MustCompile(`h1 ~ ul.hul-u`)
var suggestionMatcher = cascadia.MustCompile(`li`)

func ParseSuggestionHTML(page io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, fmt.Errorf("can not parse page: %w", err)
	}

	suggestions := doc.FindMatcher(suggestionListMatcher).
		ChildrenMatcher(suggestionMatcher).
		Map(func(i int, li *goquery.Selection) string {
			return strings.TrimSpace(li.Text())
		})
	if len(suggestions) == 0 {
		return nil, fmt.Errorf("no suggestions found")
	}
	return suggestions, nil
}
