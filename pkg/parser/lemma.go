package parser

type Lemma struct {
	Lemma          string              `json:"lemma"`
	PartOfSpeech   []string            `json:"part_of_speech,omitempty"`
	Language       string              `json:"language"`
	Transcriptions map[string][]string `json:"transcriptions,omitempty"`
	Definition     string              `json:"definition"`
	GuideWord      string              `json:"guide_word"`
	Alternative    string              `json:"alternative"`
	Grammar        []string            `json:"grammar,omitempty"`
	Examples       []string            `json:"examples,omitempty"`
}
