package querier

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/darkclainer/camgo/pkg/parser"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
)

var badgerDB struct {
	db   *badger.DB
	init sync.Once
}

func getStorage(t *testing.T) *Storage {
	db, err := getBadgerInstance()
	if err != nil {
		t.Fatalf("can not get badger instance: %v", err)
		return nil
	}
	return &Storage{
		DB: db,
	}
}

func getBadgerInstance() (*badger.DB, error) {
	var err error
	badgerDB.init.Do(func() {
		opt := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
		badgerDB.db, err = badger.Open(opt)
	})
	if err != nil {
		return nil, err
	}
	if err := badgerDB.db.DropAll(); err != nil {
		return nil, err
	}
	return badgerDB.db, nil
}

func TestMain(m *testing.M) {
	code := m.Run()
	if badgerDB.db != nil {
		_ = badgerDB.db.Close()
	}
	os.Exit(code)
}

func TestQuery(t *testing.T) {
	storage := getStorage(t)
	testCases := map[string]struct {
		lemmaID     string
		suggestions []string
		errorMsg    string
	}{
		"query": {
			lemmaID: "test_simple",
		},
		"query suggestions": {
			lemmaID:     "test_suggestions",
			suggestions: []string{"a"},
		},
		"query with error": {
			lemmaID:     "test_error",
			suggestions: []string{"a", "b"},
			errorMsg:    "hello",
		},
	}
	for name := range testCases {
		name := name
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			testStarted := time.Now()
			var queryErr error
			if tc.errorMsg != "" {
				queryErr = errors.New(tc.errorMsg)
			}
			expectedQuery := &CachedQuery{
				LemmaID:     tc.lemmaID,
				Suggestions: tc.suggestions,
				Error:       tc.errorMsg,
			}
			err := storage.PutQuery(name,
				expectedQuery.LemmaID,
				expectedQuery.Suggestions,
				queryErr,
			)
			if err != nil {
				t.Fatalf("put query failed: %v", err)
			}
			q, err := storage.GetQuery(name)
			assert.NoError(t, err)
			if !q.CreatedAt.After(testStarted) {
				t.Error("CreatedAt is before than test started")
			}
			q.CreatedAt = time.Time{}
			assert.Equal(t, expectedQuery, q)
		})
	}
}

func TestLemma(t *testing.T) {
	storage := getStorage(t)
	testCases := map[string]struct {
		lemmas   []*parser.Lemma
		errorMsg string
	}{
		"simple lemma": {
			lemmas: []*parser.Lemma{
				{
					Lemma: "hello",
				},
			},
		},
		"error lemma": {
			errorMsg: "err",
		},
	}
	for name := range testCases {
		name := name
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			testStarted := time.Now()
			var lemmaErr error
			if tc.errorMsg != "" {
				lemmaErr = errors.New(tc.errorMsg)
			}
			expectedLemma := &CachedLemma{
				Lemmas: tc.lemmas,
				Error:  tc.errorMsg,
			}
			err := storage.PutLemma(name,
				expectedLemma.Lemmas,
				lemmaErr,
			)
			if err != nil {
				t.Fatalf("put lemma failed: %v", err)
			}
			q, err := storage.GetLemma(name)
			assert.NoError(t, err)
			if !q.CreatedAt.After(testStarted) {
				t.Error("CreatedAt is before than test started")
			}
			q.CreatedAt = time.Time{}
			assert.Equal(t, expectedLemma, q)
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
