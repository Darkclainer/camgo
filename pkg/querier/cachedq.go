package querier

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"

	"github.com/darkclainer/camgo/pkg/parser"
)

type QueryInterface interface {
	GetLemma(ctx context.Context, lemmaID string) ([]*parser.Lemma, error)
	Search(ctx context.Context, query string) (string, []string, error)
	Close(ctx context.Context) error
}

type Cached struct {
	querier QueryInterface
	storage *Storage
}

func NewCached(querier QueryInterface, storage *badger.DB) *Cached {
	return &Cached{
		querier: querier,
		storage: &Storage{DB: storage},
	}
}

func (c *Cached) GetLemma(ctx context.Context, lemmaID string) ([]*parser.Lemma, error) {
	cached, err := c.storage.GetLemma(lemmaID)
	if err == nil {
		return cached.Return()
	}
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}
	lemmas, err := c.querier.GetLemma(ctx, lemmaID)
	if err := c.storage.PutLemma(lemmaID, lemmas, err); err != nil {
		// TODO: log this event
	}
	return lemmas, err
}

func (c *Cached) Search(ctx context.Context, query string) (string, []string, error) {
	cached, err := c.storage.GetQuery(query)
	if err == nil {
		return cached.Return()
	}
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return "", nil, err
	}
	// err is ErrKeyNotFound
	lemmaID, suggestions, err := c.querier.Search(ctx, query)
	if err := c.storage.PutQuery(query, lemmaID, suggestions, err); err != nil {
		// TODO: log this event
	}
	return lemmaID, suggestions, err
}

func (c *Cached) Close(ctx context.Context) error {
	var errs []error
	if closeErr := c.querier.Close(ctx); errs != nil {
		errs = append(errs, fmt.Errorf("querier close failed: %w", closeErr))
	}
	if closeErr := c.storage.Close(); errs != nil {
		errs = append(errs, fmt.Errorf("storage close failed: %w", closeErr))
	}
	if len(errs) != 0 {
		var strErrs []string
		for _, e := range errs {
			strErrs = append(strErrs, e.Error())
		}
		summary := strings.Join(strErrs, " AND ")
		return fmt.Errorf("while closing next errors happend: %s", summary)
	}
	return nil
}
