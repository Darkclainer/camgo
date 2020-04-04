package querier

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryCachedType(t *testing.T) {
	testCases := map[string]struct {
		cached   cachedQuery
		expected cachedQueryType
	}{
		"Lemma Type": {
			cached: cachedQuery{
				LemmaID: "lemma",
			},
			expected: cachedQueryLemma,
		},
		"Suggestion Type": {
			cached: cachedQuery{
				Suggestions: []string{"lemma-b"},
			},
			expected: cachedQuerySuggestion,
		},
		"Error Type": {
			cached: cachedQuery{
				Error: errors.New("test error"),
			},
			expected: cachedQueryError,
		},
		"Error Type (empty)": {
			cached:   cachedQuery{},
			expected: cachedQueryError,
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			actualType := tc.cached.Type()
			assert.Equal(t, tc.expected, actualType)
		})
	}
}

func TestCachedKeys(t *testing.T) {
	testCases := map[string]struct {
		keyType  string
		keyRaw   string
		expected []byte
		err      error
	}{
		"Cached query key": {
			keyType:  "query",
			keyRaw:   "key",
			expected: []byte{byte(queryKey), 'k', 'e', 'y'},
		},
		"Cached query empty": {
			keyType:  "query",
			keyRaw:   "",
			expected: []byte{byte(queryKey)},
		},
		"Cached lemma key": {
			keyType:  "lemma",
			keyRaw:   "key",
			expected: []byte{byte(lemmaKey), 'k', 'e', 'y'},
		},
		"Cached lemma empty": {
			keyType:  "lemma",
			keyRaw:   "",
			expected: []byte{byte(lemmaKey)},
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			var marshaledKey []byte
			var err error
			if tc.keyType == "query" { // nolint:goconst // test
				key := cachedQueryKey(tc.keyRaw)
				marshaledKey, err = key.MarshalBinary()
			} else {
				key := cachedLemmaKey(tc.keyRaw)
				marshaledKey, err = key.MarshalBinary()
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, marshaledKey)

			var unmarshaledKey []byte
			if tc.keyType == "query" {
				var ukey cachedQueryKey
				err = ukey.UnmarshalBinary(marshaledKey)
				unmarshaledKey = []byte(ukey)
			} else {
				var ukey cachedLemmaKey
				err = ukey.UnmarshalBinary(marshaledKey)
				unmarshaledKey = []byte(ukey)
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.keyRaw, string(unmarshaledKey))
		})
	}
}

func TestUnmarshalKeyError(t *testing.T) {
	testCases := map[string]struct {
		data []byte
		kt   keyType
	}{
		"Zero length": {
			data: []byte{},
		},
		"Wrong key type": {
			data: []byte{byte(queryKey), 'a'},
			kt:   lemmaKey,
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			_, err := unmarshalKey(tc.data, tc.kt)
			assert.Error(t, err)
		})
	}
}
