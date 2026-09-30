package agent

import (
	_ "embed"
	"strings"
)

// This is a transparent, Agent-reviewed interpretation set for this package,
// not a reference answer set or a claim of general natural-language coverage.
// Exact source text must match; changing a question ID never selects an answer.
//
//go:embed reviewed.json
var reviewedData []byte

type ReviewedCondition struct {
	Text     string   `json:"text"`
	Document Document `json:"document"`
}

var reviewed = func() map[string]Document {
	var entries []ReviewedCondition
	if err := DecodeJSON(reviewedData, &entries); err != nil {
		panic(err)
	}
	result := make(map[string]Document, len(entries))
	for _, entry := range entries {
		result[strings.TrimSpace(entry.Text)] = entry.Document
	}
	return result
}()

func reviewedDocument(text string) (Document, bool) {
	document, ok := reviewed[strings.TrimSpace(text)]
	return document, ok
}
