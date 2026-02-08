package skill

import (
	"regexp"
	"strings"
)

// nonAlphanumRegex matches characters that are not lowercase alphanumeric or hyphens.
var nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9-]`)

// multiHyphenRegex matches two or more consecutive hyphens.
var multiHyphenRegex = regexp.MustCompile(`-{2,}`)

// SuggestName generates a valid skill name from a free-form query string.
// It lowercases the input, replaces spaces and special characters with hyphens,
// collapses multiple hyphens, trims leading/trailing hyphens, and truncates to 64 chars.
// Returns an empty string if the result is not a valid skill name.
func SuggestName(query string) string {
	name := strings.ToLower(strings.TrimSpace(query))
	if name == "" {
		return ""
	}

	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove non-alphanumeric characters (except hyphens)
	name = nonAlphanumRegex.ReplaceAllString(name, "")

	// Collapse multiple hyphens
	name = multiHyphenRegex.ReplaceAllString(name, "-")

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Truncate to 64 chars max
	if len(name) > 64 {
		name = name[:64]
		// Don't end on a hyphen after truncation
		name = strings.TrimRight(name, "-")
	}

	// Validate the result
	if ValidateName(name) != nil {
		return ""
	}

	return name
}
