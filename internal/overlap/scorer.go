package overlap

import (
	"strings"
)

// Weights configures how similarity signals are combined into a final score.
// All weights should sum to 1.0.
type Weights struct {
	Name  float64
	Desc  float64
	Tools float64
	Body  float64
}

// OverlapWeights are used for overlap detection during skill creation and dedup checks.
var OverlapWeights = Weights{Name: 0.5, Desc: 0.3, Tools: 0.2, Body: 0.0}

// SearchWeights are used for fuzzy search and skill recommendation.
var SearchWeights = Weights{Name: 0.4, Desc: 0.4, Tools: 0.0, Body: 0.2}

// ScoreAll computes a weighted similarity score using all available signals.
// Returns a value in [0, 1].
func ScoreAll(w Weights, candidateName, candidateDesc, candidateBody string, candidateTools []string,
	existingName, existingDesc, existingBody string, existingTools []string) float64 {
	var score float64
	if w.Name > 0 {
		score += NameSimilarity(candidateName, existingName) * w.Name
	}
	if w.Desc > 0 {
		score += DescriptionSimilarity(candidateDesc, existingDesc) * w.Desc
	}
	if w.Tools > 0 {
		score += toolsOverlap(candidateTools, existingTools) * w.Tools
	}
	if w.Body > 0 {
		score += DescriptionSimilarity(candidateBody, existingBody) * w.Body
	}
	return score
}

// Score computes an overall similarity score between a candidate skill and an existing skill.
// Returns a value in [0, 1] where 1 means identical.
func Score(candidateName, candidateDesc, existingName, existingDesc string, existingTools []string) float64 {
	return ScoreAll(OverlapWeights, candidateName, candidateDesc, "", nil, existingName, existingDesc, "", existingTools)
}

// ScoreWithTools computes similarity including the candidate's allowed-tools.
func ScoreWithTools(candidateName, candidateDesc string, candidateTools []string,
	existingName, existingDesc string, existingTools []string) float64 {
	return ScoreAll(OverlapWeights, candidateName, candidateDesc, "", candidateTools, existingName, existingDesc, "", existingTools)
}

// DescriptionSimilarity computes keyword overlap between two descriptions.
// Returns a value in [0, 1].
func DescriptionSimilarity(a, b string) float64 {
	wordsA := extractKeywords(a)
	wordsB := extractKeywords(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	// Jaccard similarity on keyword sets.
	// extractKeywords already returns deduplicated keywords,
	// so we only need one map for the intersection check.
	setA := make(map[string]bool, len(wordsA))
	for _, w := range wordsA {
		setA[w] = true
	}

	intersection := 0
	for _, w := range wordsB {
		if setA[w] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
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

// extractKeywords tokenizes text and filters stop words, returning unique meaningful keywords.
func extractKeywords(text string) []string {
	text = strings.ToLower(text)

	// Replace common separators with spaces
	replacer := strings.NewReplacer(
		"-", " ", "_", " ", "/", " ", ".", " ", ",", " ",
		"(", " ", ")", " ", "[", " ", "]", " ",
	)
	text = replacer.Replace(text)

	words := strings.Fields(text)
	seen := make(map[string]bool, len(words))
	var keywords []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if stopWords[w] {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		keywords = append(keywords, w)
	}
	return keywords
}
