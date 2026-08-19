package skill

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// frontmatter is the YAML structure within --- delimiters in SKILL.md.
//
// The modeled fields are the keys skern owns. Every other key is captured by
// the inline Extra map on parse and emitted again on write, so unmodeled keys
// (top-level or under metadata) are never silently dropped — see #100.
type frontmatter struct {
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Tags         []string       `yaml:"tags,omitempty"`
	AllowedTools []string       `yaml:"allowed-tools,omitempty"`
	Metadata     Metadata       `yaml:"metadata"`
	Install      InstallConfig  `yaml:"install,omitempty"`
	Extra        map[string]any `yaml:",inline"`
}

// modeledTopLevelKeys and modeledMetadataKeys are the YAML keys skern models
// explicitly, derived from the struct tags so they cannot drift from the
// types. They guard WriteManifest against an Extra key that shadows a modeled
// field — yaml.v3 panics on that collision at marshal time.
var (
	modeledTopLevelKeys   = yamlKeys(reflect.TypeOf(frontmatter{}))
	modeledMetadataKeys   = yamlKeys(reflect.TypeOf(Metadata{}))
	modeledAuthorKeys     = yamlKeys(reflect.TypeOf(Author{}))
	modeledModifiedByKeys = yamlKeys(reflect.TypeOf(ModifiedByEntry{}))
)

// yamlKeys returns the set of YAML keys yaml.v3 would use for a struct type:
// the explicit `yaml:"<name>"` tag, or the lowercased field name when a field
// is untagged (yaml.v3's default). Inline and ignored ("-") fields are
// skipped.
func yamlKeys(t reflect.Type) map[string]bool {
	keys := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("yaml")
		name, opts, _ := strings.Cut(tag, ",")
		if name == "-" || strings.Contains(","+opts+",", ",inline,") {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		keys[name] = true
	}
	return keys
}

// ParseManifest reads a SKILL.md file and returns the parsed Skill.
func ParseManifest(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	return ParseManifestFromBytes(data)
}

// ParseManifestFromBytes parses a SKILL.md from raw bytes (for imported or
// in-memory content). ParseManifest delegates here so both paths share one
// frontmatter model.
func ParseManifestFromBytes(data []byte) (*Skill, error) {
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
		Tags:         f.Tags,
		AllowedTools: f.AllowedTools,
		Metadata:     f.Metadata,
		Install:      f.Install,
		Extra:        f.Extra,
		Body:         body,
	}, nil
}

// WriteManifest writes a Skill to a SKILL.md file.
//
// Modeled fields are written first in declaration order; unmodeled keys held
// in the Extra maps (top level, metadata, metadata.author, modified-by
// entries) follow their section, sorted by key. An Extra key that collides
// with a modeled key is an error rather than a silent overwrite.
func WriteManifest(s *Skill, path string) error {
	if err := checkExtraCollisions("frontmatter", s.Extra, modeledTopLevelKeys); err != nil {
		return err
	}
	if err := checkExtraCollisions("metadata", s.Metadata.Extra, modeledMetadataKeys); err != nil {
		return err
	}
	if err := checkExtraCollisions("metadata.author", s.Metadata.Author.Extra, modeledAuthorKeys); err != nil {
		return err
	}
	for i, m := range s.Metadata.ModifiedBy {
		if err := checkExtraCollisions(fmt.Sprintf("metadata.modified-by[%d]", i), m.Extra, modeledModifiedByKeys); err != nil {
			return err
		}
	}

	fm := frontmatter{
		Name:         s.Name,
		Description:  s.Description,
		Tags:         s.Tags,
		AllowedTools: s.AllowedTools,
		Metadata:     s.Metadata,
		Install:      s.Install,
		Extra:        s.Extra,
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

// checkExtraCollisions reports an error naming every key in extra that is
// also a modeled key in the given section.
func checkExtraCollisions(section string, extra map[string]any, modeled map[string]bool) error {
	var clashes []string
	for k := range extra {
		if modeled[k] {
			clashes = append(clashes, k)
		}
	}
	if len(clashes) == 0 {
		return nil
	}
	sort.Strings(clashes)
	return fmt.Errorf("%s: extra key(s) %s collide with skern-modeled fields; set the modeled field instead", section, strings.Join(clashes, ", "))
}

// RenderExtraValue serializes one unmodeled frontmatter value as compact YAML
// for display and comparison (diff output). Scalars render bare; mappings and
// sequences render in flow style on one line. Keys are sorted, so two
// structurally equal values always render identically.
func RenderExtraValue(v any) string {
	if v == nil {
		return ""
	}
	node := &yaml.Node{}
	if err := node.Encode(v); err != nil {
		return fmt.Sprintf("%v", v)
	}
	if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
		node.Style = yaml.FlowStyle
	}
	out, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return strings.TrimSpace(string(out))
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
