package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/darkclainer/camgo"
)

const (
	codeErrorArgs = iota + 1
	codeInternalError
)

func exitf(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

func downloadWord(query string) ([]byte, error) {
	// TODO: check query
	req, err := http.Get("https://dictionary.cambridge.org/dictionary/english/" + query)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func saveWord(path string, content []byte) {
	if err := ioutil.WriteFile(path, content, 0660); err != nil {
		exitf(codeInternalError, "can not save word to %s: %s\n",
			path,
			err.Error(),
		)
	}
}

func main() {
	webWord := flag.String("w", "", "word that you want to search in the web")
	localPath := flag.String("f", "", "Local html file for parsing")
	savePath := flag.String("s", "", "name of file where parsed content will be saved")
	flag.Parse()

	var input io.Reader
	switch {
	case *webWord != "" && *localPath != "":
		exitf(codeErrorArgs, "both -w and -f can not be specified at the same time!\n")
	case *webWord != "":
		wordBytes, err := downloadWord(*webWord)
		if err != nil {
			exitf(codeInternalError, "can not download word %s: %s\n", *webWord, err.Error())
		}
		if *savePath != "" {
			saveWord(*savePath, wordBytes)
		}
		input = ioutil.NopCloser(bytes.NewBuffer(wordBytes))
	case *localPath != "":
		file, err := os.Open(*localPath)
		if err != nil {
			exitf(codeErrorArgs, "can not open file %s: %s\n", *localPath, err.Error())
		}
		input = file
	default:
		exitf(codeErrorArgs, "you should specify either -w or -f\n")
	}

	lemmas, err := camgo.ParseLemmaHTML(input)
	if err != nil {
		exitf(codeErrorArgs, "can not parse word: %s\n", err.Error())
	}
	s, err := json.MarshalIndent(lemmas, "", "\t")
	if err != nil {
		exitf(codeErrorArgs, "can not marshal word: %s\n", err.Error())
	}
	fmt.Printf("%s\n", s)
}
