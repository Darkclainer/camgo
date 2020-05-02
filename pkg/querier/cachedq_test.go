package querier

import (
	"context"
	"errors"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/darkclainer/camgo/pkg/mocks"
	"github.com/darkclainer/camgo/pkg/parser"
)

func TestCachedGetLemma(t *testing.T) {
	storage := getStorage(t)
	expectedLemmas := []*parser.Lemma{
		{
			Lemma: "mylemma",
		},
	}
	t.Run("get through querier", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		q.On("GetLemma", mock.Anything, "test_lemma").
			Return(expectedLemmas, errors.New("test error"))
		cached := NewCached(q, storage.DB)

		lemmas, err := cached.GetLemma(context.TODO(), "test_lemma")
		q.AssertExpectations(t)
		assert.EqualError(t, err, "test error")
		assert.Equal(t, expectedLemmas, lemmas)
	})
	t.Run("get through storage", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		cached := NewCached(q, storage.DB)
		lemmas, err := cached.GetLemma(context.TODO(), "test_lemma")
		assert.EqualError(t, err, "test error")
		assert.Equal(t, expectedLemmas, lemmas)
	})
}

func TestCachedSearch(t *testing.T) {
	storage := getStorage(t)
	expected := struct {
		id          string
		suggestions []string
		err         error
	}{
		id:          "test_lemma",
		suggestions: []string{"b", "e"},
		err:         errors.New("heellooo"),
	}
	t.Run("get through querier", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		q.On("Search", mock.Anything, "test_query").
			Return(expected.id, expected.suggestions, expected.err)
		cached := NewCached(q, storage.DB)

		id, suggestions, err := cached.Search(context.TODO(), "test_query")
		q.AssertExpectations(t)
		assert.Equal(t, expected.id, id)
		assert.Equal(t, expected.suggestions, suggestions)
		assert.Equal(t, expected.err, err)
	})
	t.Run("get through cached", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		cached := NewCached(q, storage.DB)

		id, suggestions, err := cached.Search(context.TODO(), "test_query")
		q.AssertExpectations(t)
		assert.Equal(t, expected.id, id)
		assert.Equal(t, expected.suggestions, suggestions)
		assert.Equal(t, expected.err, err)
	})
}

func TestCachedClose(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(nil))
	if err != nil {
		t.Fatalf("can not open badger: %v", err)
	}
	t.Run("fine", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		q.On("Close", mock.Anything).Return(nil)
		cached := NewCached(q, db)

		err := cached.Close(context.TODO())
		q.AssertExpectations(t)
		assert.NoError(t, err)
	})
	t.Run("error in querier", func(t *testing.T) {
		q := &mocks.QueryInterface{}
		q.On("Close", mock.Anything).Return(errors.New("test err"))
		cached := NewCached(q, db)

		err := cached.Close(context.TODO())
		q.AssertExpectations(t)
		assert.Error(t, err, "test err")
	})
}
