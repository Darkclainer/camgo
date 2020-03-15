package camgo

type Lemma struct {
	Lemma          string
	PartOfSpeech   []string
	Language       string
	Transcriptions map[string][]string
	Definition     string
	GuideWord      string
	Alternative    string
	Grammar        []string
	Examples       []string
}
