package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple name", "my-skill", false},
		{"single word", "skill", false},
		{"with numbers", "skill-v2", false},
		{"numbers only", "123", false},
		{"long hyphenated", "my-really-long-skill-name", false},
		{"single char", "a", false},
		{"max length 64", "a234567890123456789012345678901234567890123456789012345678901234", false},
		// Dot-namespaced names (per #92): dots are valid segment separators.
		{"dot-namespaced", "myorg.bootstrap", false},
		{"dot and hyphen mixed", "codebase-intelligence.scan", false},
		{"multi-segment dotted", "a.b.c", false},
		{"empty", "", true},
		{"too long 65", "a2345678901234567890123456789012345678901234567890123456789012345", true},
		{"uppercase", "MySkill", true},
		{"spaces", "my skill", true},
		{"underscore", "my_skill", true},
		{"leading hyphen", "-skill", true},
		{"trailing hyphen", "skill-", true},
		{"double hyphen", "my--skill", true},
		{"special chars", "skill@name", true},
		// Dot-separator placement: same shape rules as hyphen.
		{"consecutive dots", "myorg..bootstrap", true},
		{"leading dot", ".bootstrap", true},
		{"trailing dot", "bootstrap.", true},
		{"dot then hyphen", "myorg.-foo", true},
		{"hyphen then dot", "myorg-.foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"single word", "python", false},
		{"hyphenated", "this-is-a-tag", false},
		{"numbers", "web3", false},
		{"categorical", "lang:python", false},
		{"categorical hyphenated both sides", "code-topic:code-review", false},
		{"empty", "", true},
		{"uppercase", "Featured", true},
		{"all caps", "FEATURED", true},
		{"uppercase in category", "Lang:python", true},
		{"uppercase in value", "lang:Python", true},
		{"underscore", "my_tag", true},
		{"space", "my tag", true},
		{"special chars", "c++", true},
		{"leading hyphen", "-tag", true},
		{"trailing hyphen", "tag-", true},
		{"double hyphen", "my--tag", true},
		{"two colons", "a:b:c", true},
		{"empty category", ":python", true},
		{"empty value", "lang:", true},
		{"bare colon", ":", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTag(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
