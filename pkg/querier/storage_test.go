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

func TestCachedQueryKey(t *testing.T) {
	testCases := map[string]struct {
		keyRaw   string
		expected []byte
		err      error
	}{
		"Cached query key": {
			keyRaw:   "key",
			expected: []byte{byte(queryKey), 'k', 'e', 'y'},
		},
		"Cached query empty": {
			keyRaw:   "",
			expected: []byte{byte(queryKey)},
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			key := cachedQueryKey(tc.keyRaw)
			marshaledKey, err := key.MarshalBinary()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, marshaledKey)

			var unmarshaledKey cachedQueryKey
			err = unmarshaledKey.UnmarshalBinary(marshaledKey)
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
		"Zero lenght": {
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

func TestCachedLemmaKey(t *testing.T) {
	testCases := map[string]struct {
		keyRaw   string
		expected []byte
		err      error
	}{
		"Cached query key": {
			keyRaw:   "key",
			expected: []byte{byte(lemmaKey), 'k', 'e', 'y'},
		},
		"Cached query empty": {
			keyRaw:   "",
			expected: []byte{byte(lemmaKey)},
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			key := cachedLemmaKey(tc.keyRaw)
			marshaledKey, err := key.MarshalBinary()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, marshaledKey)

			var unmarshaledKey cachedLemmaKey
			err = unmarshaledKey.UnmarshalBinary(marshaledKey)
			assert.NoError(t, err)
			assert.Equal(t, tc.keyRaw, string(unmarshaledKey))
		})
	}
}
