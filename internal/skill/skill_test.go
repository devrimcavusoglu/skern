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
		{"empty", "", true},
		{"too long 65", "a2345678901234567890123456789012345678901234567890123456789012345", true},
		{"uppercase", "MySkill", true},
		{"spaces", "my skill", true},
		{"underscore", "my_skill", true},
		{"leading hyphen", "-skill", true},
		{"trailing hyphen", "skill-", true},
		{"double hyphen", "my--skill", true},
		{"special chars", "skill@name", true},
		{"dot", "my.skill", true},
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
