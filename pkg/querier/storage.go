package querier

import (
	"errors"
	"fmt"

	"github.com/darkclainer/camgo/pkg/parser"
)

type keyType byte

const (
	queryKey keyType = iota + 1
	lemmaKey
)

type cachedQueryType int

const (
	cachedQueryLemma cachedQueryType = iota + 1
	cachedQuerySuggestion
	cachedQueryError
)

type cachedQuery struct {
	LemmaID     string
	Suggestions []string
	Error       error
}

func (cq *cachedQuery) Type() cachedQueryType {
	switch {
	case cq.LemmaID != "":
		return cachedQueryLemma
	case len(cq.Suggestions) != 0:
		return cachedQuerySuggestion
	default:
		return cachedQueryError
	}
}

type cachedQueryKey string

func (k cachedQueryKey) MarshalBinary() ([]byte, error) {
	return marshalKey(string(k), queryKey)
}

func (k *cachedQueryKey) UnmarshalBinary(data []byte) error {
	unmarshalledKey, err := unmarshalKey(data, queryKey)
	if err != nil {
		return err
	}
	*k = cachedQueryKey(unmarshalledKey)
	return nil
}

type cachedLemma struct {
	Lemmas []*parser.Lemma
	Error  error
}

type cachedLemmaKey string

func (k cachedLemmaKey) MarshalBinary() ([]byte, error) {
	return marshalKey(string(k), lemmaKey)
}

func (k *cachedLemmaKey) UnmarshalBinary(data []byte) error {
	unmarshalledKey, err := unmarshalKey(data, lemmaKey)
	if err != nil {
		return err
	}
	*k = cachedLemmaKey(unmarshalledKey)
	return nil
}

func marshalKey(k string, t keyType) ([]byte, error) {
	result := make([]byte, 0, len(k)+1)
	result = append(result, byte(t))
	return append(result, []byte(k)...), nil
}

func unmarshalKey(data []byte, expected keyType) (string, error) {
	if len(data) < 1 {
		return "", errors.New("key lenght must be at least 1")
	}
	if data[0] != byte(expected) {
		return "", fmt.Errorf("key type doesn't equal to expected type")
	}
	return string(data[1:]), nil
}
