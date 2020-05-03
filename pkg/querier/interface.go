package querier

import (
	"context"

	"github.com/darkclainer/camgo/pkg/parser"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name Querier -output ../mocks/

type Querier interface {
	GetLemma(ctx context.Context, lemmaID string) ([]*parser.Lemma, error)
	Search(ctx context.Context, query string) (string, []string, error)
	Close(ctx context.Context) error
}
