package querier

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v2"

	"github.com/darkclainer/camgo/pkg/parser"
)

const ttlForErros = time.Hour * 24

type Storage struct {
	DB *badger.DB
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) GetQuery(query string) (*cachedQuery, error) {
	key := marshalKey(query, queryKey)
	var queryValue cachedQuery
	err := s.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(value []byte) error {
			return json.Unmarshal(value, &queryValue)
		})

	})
	if err != nil {
		return nil, err
	}
	return &queryValue, nil
}
func (s *Storage) PutQuery(query string, lemmaID string, suggestions []string, queryErr error) error {
	key := marshalKey(query, queryKey)
	var errString string
	if queryErr != nil {
		errString = queryErr.Error()
	}
	value := cachedQuery{
		LemmaID:     lemmaID,
		Suggestions: suggestions,
		Error:       errString,
		CreatedAt:   time.Now(),
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.DB.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, data)
		if queryErr != nil {
			entry = entry.WithTTL(ttlForErros)
		}
		return txn.SetEntry(entry)
	})
}
func (s *Storage) GetLemma(lemmaID string) (*cachedLemma, error) {
	key := marshalKey(lemmaID, lemmaKey)
	var lemmaValue cachedLemma
	err := s.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(value []byte) error {
			return json.Unmarshal(value, &lemmaValue)
		})

	})
	if err != nil {
		return nil, err
	}
	return &lemmaValue, nil
}
func (s *Storage) PutLemma(lemmaID string, lemmas []*parser.Lemma, lemmaErr error) error {
	var errString string
	if lemmaErr != nil {
		errString = lemmaErr.Error()
	}
	key := marshalKey(lemmaID, lemmaKey)
	value := cachedLemma{
		Lemmas:    lemmas,
		Error:     errString,
		CreatedAt: time.Now(),
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.DB.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, data)
		if lemmaErr != nil {
			entry = entry.WithTTL(ttlForErros)
		}
		return txn.SetEntry(entry)
	})
}

type keyType byte

const (
	queryKey keyType = iota + 1
	lemmaKey
)

type cachedQuery struct {
	LemmaID     string
	Suggestions []string
	Error       string
	CreatedAt   time.Time
}

func (cq *cachedQuery) Return() (string, []string, error) {
	return cq.LemmaID, cq.Suggestions, errors.New(cq.Error)
}

type cachedLemma struct {
	Lemmas    []*parser.Lemma
	Error     string
	CreatedAt time.Time
}

func (cl *cachedLemma) Return() ([]*parser.Lemma, error) {
	return cl.Lemmas, errors.New(cl.Error)
}

func marshalKey(k string, t keyType) []byte {
	result := make([]byte, 0, len(k)+1) // nolint:gomnd // next line we will apend 1 byte
	result = append(result, byte(t))
	return append(result, []byte(k)...)
}
