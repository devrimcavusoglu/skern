package skill

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// frontmatter is the YAML structure within --- delimiters in SKILL.md.
type frontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	AllowedTools []string `yaml:"allowed-tools,omitempty"`
	Metadata     Metadata `yaml:"metadata"`
}

// ParseManifest reads a SKILL.md file and returns the parsed Skill.
func ParseManifest(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	content := string(data)
	if len(strings.TrimSpace(content)) == 0 {
		return nil, fmt.Errorf("manifest file is empty")
	}

	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var f frontmatter
	if err := yaml.Unmarshal([]byte(fm), &f); err != nil {
		return nil, fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	return &Skill{
		Name:         f.Name,
		Description:  f.Description,
		AllowedTools: f.AllowedTools,
		Metadata:     f.Metadata,
		Body:         body,
	}, nil
}

// WriteManifest writes a Skill to a SKILL.md file.
func WriteManifest(s *Skill, path string) error {
	fm := frontmatter{
		Name:         s.Name,
		Description:  s.Description,
		AllowedTools: s.AllowedTools,
		Metadata:     s.Metadata,
	}

	yamlBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return fmt.Errorf("marshaling YAML frontmatter: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(s.Body)

	return os.WriteFile(path, []byte(buf.String()), 0o644)
}

// splitFrontmatter splits a SKILL.md into YAML frontmatter and body.
func splitFrontmatter(content string) (string, string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return "", "", fmt.Errorf("manifest must start with --- delimiter")
	}

	rest := content[4:] // skip opening ---\n
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		// Check for --- at end of file without trailing newline after body
		if strings.HasSuffix(rest, "\n---") {
			fm := rest[:len(rest)-4]
			return fm, "", nil
		}
		return "", "", fmt.Errorf("missing closing --- delimiter in manifest")
	}

	fm := rest[:idx]
	body := strings.TrimPrefix(rest[idx+5:], "\n") // skip \n---\n and optional leading newline

	return fm, body, nil
}
