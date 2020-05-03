package querier

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"

	"github.com/darkclainer/camgo/pkg/parser"
)

type Cached struct {
	querier Querier
	storage *Storage
}

func NewCached(querier Querier, storage *badger.DB) *Cached {
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
	if dbErr := c.storage.PutLemma(lemmaID, lemmas, err); dbErr != nil { // nolint:staticcheck // todo
		// TODO: log this event
	}
	return lemmas, err
}

func (c *Cached) Search(ctx context.Context, query string) (lemmaID string, suggestions []string, err error) {
	cached, err := c.storage.GetQuery(query)
	if err == nil {
		return cached.Return()
	}
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return "", nil, err
	}
	// err is ErrKeyNotFound
	lemmaID, suggestions, err = c.querier.Search(ctx, query)
	if dbErr := c.storage.PutQuery(query, lemmaID, suggestions, err); dbErr != nil { // nolint:staticcheck // todo
		// TODO: log this event
	}
	return lemmaID, suggestions, err
}

func (c *Cached) Close(ctx context.Context) error {
	var errs []error
	if closeErr := c.querier.Close(ctx); closeErr != nil {
		errs = append(errs, fmt.Errorf("querier close failed: %w", closeErr))
	}
	if closeErr := c.storage.Close(); closeErr != nil {
		errs = append(errs, fmt.Errorf("storage close failed: %w", closeErr))
	}
	if len(errs) != 0 {
		var strErrs []string
		for _, e := range errs {
			strErrs = append(strErrs, e.Error())
		}
		summary := strings.Join(strErrs, " AND ")
		return fmt.Errorf("while closing next errors happened: %s", summary)
	}
	return nil
}
