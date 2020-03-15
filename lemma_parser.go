package camgo

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

func debugPrint(v interface{}) {
	s, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	fmt.Printf("%s\n", s)
}

var _ = goquery.NewDocumentFromReader

type enrichFunc func(lctx *Lemma, sel *goquery.Selection) ([]*Lemma, error)

func enrichLemmas(lctx *Lemma, sel *goquery.Selection, f enrichFunc) ([]*Lemma, error) {
	var lastError error
	lemmas := make([]*Lemma, 0, sel.Length())
	sel.EachWithBreak(func(i int, dictionary *goquery.Selection) bool {
		newLemmaContext := *lctx
		newLemmas, err := f(&newLemmaContext, dictionary)
		if err != nil {
			lastError = err
			return false
		}
		lemmas = append(lemmas, newLemmas...)
		return true
	})
	return lemmas, lastError
}

var dictionaryMatcher = cascadia.MustCompile(`div[class*="dictionary"][data-id]`)

func ParseLemmaHTML(page io.Reader) ([]*Lemma, error) {
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, fmt.Errorf("can not parse page: %w", err)
	}

	dictionaries := doc.FindMatcher(dictionaryMatcher)
	lemmas, err := enrichLemmas(new(Lemma), dictionaries, parseDictionary)
	debugPrint(lemmas)
	return lemmas, err
}

var dataIDToLanguage = map[string]string{
	"unknown": "unknown",
	"cald4":   "british",
	"cacd":    "american-english",
	"cbed":    "business-english",
}

func getLanguageFromDataID(dictionary *goquery.Selection) (string, error) {
	dataID := dictionary.AttrOr("data-id", "unknown")
	language, ok := dataIDToLanguage[dataID]
	if !ok {
		return "", fmt.Errorf("div.dictionary has unknown data-id attr: %s", dataID)
	}
	return language, nil
}

var dictionaryEntryMatcher = cascadia.MustCompile(strings.Join([]string{
	`div[class*="pos-header"][class*="dpos-h"]`,
	`div[class="pv-block"]`,
	`div[class="idiom-block"]`,
}, ", "))

func parseDictionary(lctx *Lemma, dictionary *goquery.Selection) ([]*Lemma, error) {
	language, err := getLanguageFromDataID(dictionary)
	if err != nil {
		return nil, err
	}
	lctx.Language = language
	/*
		dictionary can have three different type of "lemmas":
		1. div[class*="pos-header"][class*="dpos-h"] for simple words like "ghost"
		2. div[class="pv-block"] for phrasal verbs
		3. div[class="idiom-block"] for idioms
	*/
	entries := dictionary.FindMatcher(dictionaryEntryMatcher)
	return enrichLemmas(lctx, entries, func(lctx *Lemma, sel *goquery.Selection) ([]*Lemma, error) {
		switch {
		case sel.HasClass("pos-header"):
			return parsePosHeader(lctx, sel)
		case sel.HasClass("pv-block"):
			return parsePVBlock(lctx, sel)
		case sel.HasClass("idiom-block"):
			return parseIdiomBlock(lctx, sel)
		default:
			panic("Uknown dictionary entry: " + sel.AttrOr("class", ""))
		}
	})
}

var dsenseMatcher = cascadia.MustCompile(`div.dsense`)

func parsePosHeader(lctx *Lemma, posHeader *goquery.Selection) ([]*Lemma, error) {
	if err := updateLemmaWithHeader(lctx, posHeader); err != nil {
		return nil, err
	}
	dsenses := posHeader.Next().ChildrenMatcher(dsenseMatcher)
	return enrichLemmas(lctx, dsenses, parseDSense)
}

var headwordMatcher = cascadia.MustCompile(`span[class^=headword]`)

var posgramMatcher = cascadia.MustCompile(`div[class^=posgram]`)

func updateLemmaWithHeader(lctx *Lemma, header *goquery.Selection) error {
	headword := header.FindMatcher(headwordMatcher)
	if headword.Length() == 0 {
		return fmt.Errorf(".pos-header has not .headword elements")
	}
	lctx.Lemma = strings.TrimSpace(headword.Text())

	if err := updateLemmaWithPosgram(lctx, header.ChildrenMatcher(posgramMatcher)); err != nil {
		return err
	}
	lctx.Transcriptions = getTranscriptions(header)
	return nil
}

func updateLemmaWithPosgram(lctx *Lemma, posgram *goquery.Selection) error {
	lctx.PartOfSpeech = getPartOfSpeech(posgram)
	lctx.Grammar = getGrammar(posgram)
	return nil
}

var grammarBlockMatcher = cascadia.MustCompile(`span[class^=gram]`)
var grammarMatcher = cascadia.MustCompile(`span[class^=gc]`)

// getGrammar extracts grammar from children span.gram elements
func getGrammar(sel *goquery.Selection) []string {
	gram := sel.
		ChildrenMatcher(grammarBlockMatcher).
		FindMatcher(grammarMatcher)
	grammar := gram.Map(func(i int, sel *goquery.Selection) string {
		return sel.Text()
	})
	sort.Strings(grammar)
	return grammar
}

var partOfSpeechMatcher = cascadia.MustCompile(`span[class^=pos]`)

// getPartOfSpeech extracts POS from childrens span.pos elements
func getPartOfSpeech(sel *goquery.Selection) []string {
	pos := sel.ChildrenMatcher(partOfSpeechMatcher)
	partOfSpeech := pos.Map(func(i int, sel *goquery.Selection) string {
		return sel.Text()
	})
	sort.Strings(partOfSpeech)
	return partOfSpeech
}

var transcriptionFullMatcher = cascadia.MustCompile(`span[class*="dpron-i"]`)
var transcriptionRegionMatcher = cascadia.MustCompile(`span[class^="region"]`)
var transcriptionIPAMatcher = cascadia.MustCompile(`span[class^="ipa"]`)

// getTranscriptions extracts transcriptions from children of sel
func getTranscriptions(sel *goquery.Selection) map[string][]string {
	dprons := sel.ChildrenMatcher(transcriptionFullMatcher)
	transcriptions := make(map[string][]string, dprons.Length())
	dprons.Each(func(i int, dpron *goquery.Selection) {
		region := dpron.ChildrenMatcher(transcriptionRegionMatcher).First().Text()
		ipas := dpron.FindMatcher(transcriptionIPAMatcher).Map(transfromIPAToString)
		transcriptions[region] = ipas
	})
	return transcriptions
}

func transfromIPAToString(i int, ipa *goquery.Selection) string {
	var parts []string
	ipa.Contents().Each(func(i int, sel *goquery.Selection) {
		switch goquery.NodeName(sel) {
		case "#text":
			parts = append(parts, sel.Text())
		case "span":
			parts = append(parts, transformIPASpan(sel))
		}
	})
	return strings.Join(parts, "")
}

func transformIPASpan(ipaSpan *goquery.Selection) string {
	return IpaSuperscript(ipaSpan.Text())
}

var dsensehMatcher = cascadia.MustCompile(`h3.dsense_h`)

var dsenseEntryMatcher = cascadia.MustCompile(strings.Join([]string{
	`div.def-block`,
	`div.phrase-block`,
}, ", "))

func parseDSense(lctx *Lemma, dsense *goquery.Selection) ([]*Lemma, error) {
	lctx.GuideWord = getGuideWordFromDSenseH(dsense.ChildrenMatcher(dsensehMatcher))
	// TODO: get pos here

	entries := dsense.FindMatcher(dsenseEntryMatcher)
	return enrichLemmas(lctx, entries, func(lctx *Lemma, sel *goquery.Selection) ([]*Lemma, error) {
		switch {
		case sel.HasClass("def-block"):
			return parseDefBlock(lctx, sel)
		case sel.HasClass("phrase-block"):
			return parsePhraseBlock(lctx, sel)
		default:
			panic("Uknown dsense entry: " + sel.AttrOr("class", ""))
		}
	})
}

var guidewordMatcher = cascadia.MustCompile(`span.guideword`)

func getGuideWordFromDSenseH(dsenseh *goquery.Selection) string {
	guideword := dsenseh.ChildrenMatcher(guidewordMatcher)
	return strings.ToLower(guideword.Children().Text())
}

var ddefhMatcher = cascadia.MustCompile(`div.ddef_h`)
var defInfoMatcher = cascadia.MustCompile(`span.def-info`)
var defBodyMatcher = cascadia.MustCompile(`div.def-body`)

func parseDefBlock(lctx *Lemma, defBlock *goquery.Selection) ([]*Lemma, error) {
	ddefh := defBlock.ChildrenMatcher(ddefhMatcher)
	definition, err := getDefinitionFromDDefH(ddefh)
	if err != nil {
		return nil, err
	}
	lctx.Definition = definition
	defInfo := ddefh.ChildrenMatcher(defInfoMatcher)
	if grammar := getGrammar(defInfo); len(grammar) != 0 {
		lctx.Grammar = grammar
	}
	lctx.Alternative = getAlternativeForm(defInfo)
	lctx.Examples = getExamples(defBlock.ChildrenMatcher(defBodyMatcher))
	return []*Lemma{lctx}, nil
}

var defMatcher = cascadia.MustCompile(`div.def`)

func getDefinitionFromDDefH(ddefh *goquery.Selection) (string, error) {
	def := ddefh.ChildrenMatcher(defMatcher)
	if def.Length() == 0 {
		return "", fmt.Errorf("div.ddef_h has no div.dev")
	}
	definition := strings.TrimSpace(def.Text())
	definition = strings.TrimSuffix(definition, ":")
	return definition, nil
}

var alternativeMatcher = cascadia.MustCompile(`span.v`)

func getAlternativeForm(defInfo *goquery.Selection) string {
	return defInfo.FindMatcher(alternativeMatcher).Text()
}

var exampleMatcher = cascadia.MustCompile(`div.examp`)

func getExamples(defBody *goquery.Selection) []string {
	return defBody.ChildrenMatcher(exampleMatcher).Map(func(i int, examp *goquery.Selection) string {
		return strings.TrimSpace(examp.Text())
	})
}

func parsePhraseBlock(lctx *Lemma, dictionary *goquery.Selection) ([]*Lemma, error) {
	println("phrase-block")
	return nil, nil
}

func parsePVBlock(lctx *Lemma, dictionary *goquery.Selection) ([]*Lemma, error) {
	println("pv-block")
	return nil, nil
}
func parseIdiomBlock(lctx *Lemma, dictionary *goquery.Selection) ([]*Lemma, error) {
	println("idiom-block")
	return nil, nil
}
