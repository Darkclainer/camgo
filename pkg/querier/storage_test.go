package querier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			var binaryKey []byte
			var err error
			if tc.keyType == "query" { // nolint:goconst // test
				binaryKey = marshalKey(tc.keyRaw, queryKey)
			} else {
				binaryKey = marshalKey(tc.keyRaw, lemmaKey)
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, binaryKey)
		})
	}
}
