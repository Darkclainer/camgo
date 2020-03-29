package camgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type JSONCase struct {
	Name    string `json:"-"`
	Path    string `json:"-"`
	Length  int
	Type    string
	Content *json.RawMessage
}

type LemmaCase struct {
	Index int
	Lemma *Lemma
}

type LemmaTest struct {
	Name        string
	Length      int
	Cases       []*LemmaCase
	HTMLContent []byte
}

type SuggestionTest struct {
	Expected    []string
	HTMLContent []byte
}

type TestCases struct {
	LemmaTests      []*LemmaTest
	SuggestionTests []*SuggestionTest
}

func loadTestCases(jsonPath, htmlPath string) (*TestCases, error) {
	jsonFiles, err := filepath.Glob(filepath.Join(jsonPath, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("no json files was found in: %s", jsonPath)
	}
	jsonCases, err := loadJSONCases(jsonFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to load json cases: %w", err)
	}

	var lemmaTests []*LemmaTest
	var suggestionTests []*SuggestionTest
	for _, jsonCase := range jsonCases {
		switch jsonCase.Type {
		case "lemmas":
			var lemmaCases []*LemmaCase
			if err := json.Unmarshal(*jsonCase.Content, &lemmaCases); err != nil {
				return nil, fmt.Errorf("can not parse lemma test cases for '%s': %w", jsonCase.Name, err)
			}
			htmlContent, err := loadHTMLContent(htmlPath, jsonCase.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to load html content for test '%s': %w", jsonCase.Name, err)
			}
			lemmaTests = append(lemmaTests, &LemmaTest{
				Name:        jsonCase.Name,
				Length:      jsonCase.Length,
				Cases:       lemmaCases,
				HTMLContent: htmlContent,
			})
		case "suggestions":
			_ = suggestionTests
			panic("suggestion unimplemented yet")
		default:
			return nil, fmt.Errorf("unknown test type for file '%s': %s", jsonCase.Path, jsonCase.Type)
		}
	}
	return &TestCases{
		LemmaTests:      lemmaTests,
		SuggestionTests: suggestionTests,
	}, nil
}

func loadHTMLContent(htmlPath, name string) ([]byte, error) {
	path := filepath.Join(htmlPath, name+".html")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read html file '%s': %w", path, err)
	}
	return content, nil
}

func loadJSONCases(paths []string) ([]*JSONCase, error) {
	var jsonCases []*JSONCase
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("can open file '%s': %w", path, err)
		}
		var jsonCase JSONCase
		if err := json.NewDecoder(file).Decode(&jsonCase); err != nil {
			return nil, fmt.Errorf("can not decode file '%s': %w", path, err)
		}
		jsonCase.Path = path
		jsonCase.Name = strings.TrimSuffix(filepath.Base(path), ".json")
		jsonCases = append(jsonCases, &jsonCase)
	}
	return jsonCases, nil
}

var LoadedCases *TestCases

func TestMain(m *testing.M) {
	cases, err := loadTestCases("testdata/json", "testdata/html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can not load test files: %s\n", err)
		os.Exit(1) //nolint:gomnd // status codes is common knowledge
	}
	LoadedCases = cases
	os.Exit(m.Run())
}

func TestLemmaParser(t *testing.T) {
	testCases := LoadedCases.LemmaTests
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			parsedLemmas, err := ParseLemmaHTML(bytes.NewReader(tc.HTMLContent))
			if err != nil {
				t.Errorf("Can not parse html content: %s", err)
				t.FailNow()
			}
			if !assert.Equal(t, tc.Length, len(parsedLemmas)) {
				t.FailNow()
			}
			for _, lemmaCase := range tc.Cases {
				assert.Equalf(t, lemmaCase.Lemma, parsedLemmas[lemmaCase.Index], "Lemmas at index %d doesn't match", lemmaCase.Index)
			}
		})
	}
}
