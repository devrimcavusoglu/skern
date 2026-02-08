package overlap

import (
	"strings"
)

// Weight constants for combining similarity signals.
const (
	nameWeight        = 0.5
	descriptionWeight = 0.3
	toolsWeight       = 0.2
)

// Score computes an overall similarity score between a candidate skill and an existing skill.
// Returns a value in [0, 1] where 1 means identical.
func Score(candidateName, candidateDesc, existingName, existingDesc string, existingTools []string) float64 {
	nameSim := NameSimilarity(candidateName, existingName)
	descSim := DescriptionSimilarity(candidateDesc, existingDesc)
	toolSim := toolsOverlap(nil, existingTools) // candidate has no tools at creation time

	return nameSim*nameWeight + descSim*descriptionWeight + toolSim*toolsWeight
}

// ScoreWithTools computes similarity including the candidate's allowed-tools.
func ScoreWithTools(candidateName, candidateDesc string, candidateTools []string,
	existingName, existingDesc string, existingTools []string) float64 {
	nameSim := NameSimilarity(candidateName, existingName)
	descSim := DescriptionSimilarity(candidateDesc, existingDesc)
	toolSim := toolsOverlap(candidateTools, existingTools)

	return nameSim*nameWeight + descSim*descriptionWeight + toolSim*toolsWeight
}

// DescriptionSimilarity computes keyword overlap between two descriptions.
// Returns a value in [0, 1].
func DescriptionSimilarity(a, b string) float64 {
	wordsA := extractKeywords(a)
	wordsB := extractKeywords(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	// Jaccard similarity on keyword sets
	setA := make(map[string]bool, len(wordsA))
	for _, w := range wordsA {
		setA[w] = true
	}

	setB := make(map[string]bool, len(wordsB))
	for _, w := range wordsB {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// toolsOverlap computes the overlap between two allowed-tools lists.
// Returns a value in [0, 1].
func toolsOverlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	setA := make(map[string]bool, len(a))
	for _, t := range a {
		setA[strings.ToLower(strings.TrimSpace(t))] = true
	}

	intersection := 0
	for _, t := range b {
		if setA[strings.ToLower(strings.TrimSpace(t))] {
			intersection++
		}
	}

	union := len(setA) + len(b) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// stopWords are common English words filtered out during keyword extraction.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true,
	"is": true, "are": true, "was": true, "were": true, "be": true,
	"been": true, "being": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "shall": true,
	"can": true, "to": true, "of": true, "in": true, "for": true,
	"on": true, "with": true, "at": true, "by": true, "from": true,
	"as": true, "into": true, "about": true, "it": true, "its": true,
	"this": true, "that": true, "these": true, "those": true,
	"not": true, "but": true, "if": true, "then": true, "so": true,
	"than": true, "too": true, "very": true, "just": true,
}

// extractKeywords tokenizes text and filters stop words, returning meaningful keywords.
func extractKeywords(text string) []string {
	text = strings.ToLower(text)

	// Replace common separators with spaces
	replacer := strings.NewReplacer(
		"-", " ", "_", " ", "/", " ", ".", " ", ",", " ",
		"(", " ", ")", " ", "[", " ", "]", " ",
	)
	text = replacer.Replace(text)

	words := strings.Fields(text)
	var keywords []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if stopWords[w] {
			continue
		}
		keywords = append(keywords, w)
	}
	return keywords
}
