// Package overlap provides fuzzy matching and similarity scoring for skill deduplication.
package overlap

import (
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/skill"
)

const (
	// WarnThreshold is the similarity score at which overlap warnings are shown.
	WarnThreshold = 0.6
	// BlockThreshold is the similarity score at which creation is blocked (requires --force).
	BlockThreshold = 0.9
)

// Match represents a detected overlap between a new skill and an existing one.
type Match struct {
	Name  string      `json:"name"`
	Scope skill.Scope `json:"scope"`
	Score float64     `json:"score"`
}

// Check evaluates a candidate skill name and description against a set of existing skills,
// returning any matches above the warn threshold.
func Check(name, description string, existing []skill.Skill, scopes []skill.Scope) []Match {
	var matches []Match

	for i, s := range existing {
		score := Score(name, description, s.Name, s.Description, s.AllowedTools)
		if score >= WarnThreshold {
			scope := skill.ScopeUser
			if i < len(scopes) {
				scope = scopes[i]
			}
			matches = append(matches, Match{
				Name:  s.Name,
				Scope: scope,
				Score: score,
			})
		}
	}

	return matches
}

// MaxScore returns the highest score from a set of matches.
func MaxScore(matches []Match) float64 {
	max := 0.0
	for _, m := range matches {
		if m.Score > max {
			max = m.Score
		}
	}
	return max
}

// ShouldBlock returns true if any match meets or exceeds the block threshold.
func ShouldBlock(matches []Match) bool {
	return MaxScore(matches) >= BlockThreshold
}

// levenshtein computes the Levenshtein edit distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Use single-row DP for space efficiency.
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev = curr
	}

	return prev[len(b)]
}

// NameSimilarity computes similarity between two skill names using
// Levenshtein distance normalized to [0, 1], with bonuses for prefix/suffix overlap.
func NameSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 0.0
	}

	if a == b {
		return 1.0
	}

	// Levenshtein-based similarity
	dist := levenshtein(a, b)
	levSim := 1.0 - float64(dist)/float64(maxLen)

	// Prefix/suffix bonus
	bonus := 0.0
	prefixLen := commonPrefixLen(a, b)
	suffixLen := commonSuffixLen(a, b)

	if prefixLen > 0 {
		bonus += float64(prefixLen) / float64(maxLen) * 0.15
	}
	if suffixLen > 0 {
		bonus += float64(suffixLen) / float64(maxLen) * 0.1
	}

	// Containment check: if one name contains the other
	if strings.Contains(a, b) || strings.Contains(b, a) {
		minLen := min(len(a), len(b))
		bonus += float64(minLen) / float64(maxLen) * 0.2
	}

	result := levSim + bonus
	if result > 1.0 {
		result = 1.0
	}
	return result
}

func commonPrefixLen(a, b string) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func commonSuffixLen(a, b string) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[len(a)-1-i] != b[len(b)-1-i] {
			return i
		}
	}
	return n
}
