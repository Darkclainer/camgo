package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/darkclainer/camgo/pkg/querier"
)

const (
	codeErrorArgs = iota + 1
	codeNotFound
	codeInternalError
	timeoutSecounds = 10
)

func exitf(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

func main() {
	query := flag.String("q", "", "query that you want to search in the web")
	flag.Parse()

	if *query == "" {
		exitf(codeErrorArgs, "you should specify arguments\n")
	}
	q := querier.NewQuerier(nil, &querier.QuerierConfig{
		ExtraHeader: map[string]string{
			"User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0",
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*timeoutSecounds)
	defer cancel()

	lemmaID, err := q.Search(ctx, *query)
	if err != nil {
		var suggestions querier.ErrLemmaNotFound
		if !errors.As(err, &suggestions) {
			exitf(codeInternalError, "unknown error: %s\n", err)
		}
		exitf(codeNotFound, "May be you mean:\n%s\n", strings.Join(suggestions, "\n"))
	}
	lemmas, err := q.GetLemma(ctx, lemmaID)
	if err != nil {
		exitf(codeInternalError, "unknown error: %s\n", err)
	}
	s, err := json.MarshalIndent(lemmas, "", "\t")
	if err != nil {
		exitf(codeErrorArgs, "can not marshal word: %s\n", err.Error())
	}
	fmt.Printf("%s\n", s)
}
