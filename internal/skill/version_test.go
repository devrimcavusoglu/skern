package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "valid version",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "default version",
			input: "0.1.0",
			want:  Version{Major: 0, Minor: 1, Patch: 0},
		},
		{
			name:  "large numbers",
			input: "10.20.30",
			want:  Version{Major: 10, Minor: 20, Patch: 30},
		},
		{
			name:    "too few parts",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "too many parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "non-numeric major",
			input:   "x.1.0",
			wantErr: true,
		},
		{
			name:    "non-numeric minor",
			input:   "1.x.0",
			wantErr: true,
		},
		{
			name:    "non-numeric patch",
			input:   "1.0.x",
			wantErr: true,
		},
		{
			name:    "negative major",
			input:   "-1.0.0",
			wantErr: true,
		},
		{
			name:    "negative minor",
			input:   "1.-1.0",
			wantErr: true,
		},
		{
			name:    "negative patch",
			input:   "1.0.-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{
			name: "simple version",
			v:    Version{Major: 1, Minor: 2, Patch: 3},
			want: "1.2.3",
		},
		{
			name: "zero version",
			v:    Version{Major: 0, Minor: 0, Patch: 0},
			want: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.v.String())
		})
	}
}

func TestVersion_BumpPatch(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	bumped := v.BumpPatch()
	assert.Equal(t, Version{Major: 1, Minor: 2, Patch: 4}, bumped)
}

func TestVersion_BumpMinor(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	bumped := v.BumpMinor()
	assert.Equal(t, Version{Major: 1, Minor: 3, Patch: 0}, bumped)
}

func TestVersion_BumpMajor(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	bumped := v.BumpMajor()
	assert.Equal(t, Version{Major: 2, Minor: 0, Patch: 0}, bumped)
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		level   string
		want    string
		wantErr bool
	}{
		{
			name:    "bump patch",
			version: "1.2.3",
			level:   "patch",
			want:    "1.2.4",
		},
		{
			name:    "bump minor",
			version: "1.2.3",
			level:   "minor",
			want:    "1.3.0",
		},
		{
			name:    "bump major",
			version: "1.2.3",
			level:   "major",
			want:    "2.0.0",
		},
		{
			name:    "bump default version patch",
			version: "0.1.0",
			level:   "patch",
			want:    "0.1.1",
		},
		{
			name:    "bump default version minor",
			version: "0.1.0",
			level:   "minor",
			want:    "0.2.0",
		},
		{
			name:    "bump default version major",
			version: "0.1.0",
			level:   "major",
			want:    "1.0.0",
		},
		{
			name:    "invalid version",
			version: "bad",
			level:   "patch",
			wantErr: true,
		},
		{
			name:    "invalid level",
			version: "1.0.0",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpVersion(tt.version, tt.level)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		wantKind string
		wantNew  bool
		wantErr  bool
	}{
		{
			name:     "equal versions",
			a:        "1.0.0",
			b:        "1.0.0",
			wantKind: "",
			wantNew:  false,
		},
		{
			name:     "patch upgrade",
			a:        "1.0.0",
			b:        "1.0.1",
			wantKind: "patch",
			wantNew:  true,
		},
		{
			name:     "minor upgrade",
			a:        "1.0.0",
			b:        "1.1.0",
			wantKind: "minor",
			wantNew:  true,
		},
		{
			name:     "major upgrade",
			a:        "1.0.0",
			b:        "2.0.0",
			wantKind: "major",
			wantNew:  true,
		},
		{
			name:     "patch downgrade",
			a:        "1.0.1",
			b:        "1.0.0",
			wantKind: "patch",
			wantNew:  false,
		},
		{
			name:     "minor downgrade",
			a:        "1.1.0",
			b:        "1.0.0",
			wantKind: "minor",
			wantNew:  false,
		},
		{
			name:     "major downgrade",
			a:        "2.0.0",
			b:        "1.0.0",
			wantKind: "major",
			wantNew:  false,
		},
		{
			name:     "major takes precedence over minor",
			a:        "1.5.0",
			b:        "2.0.0",
			wantKind: "major",
			wantNew:  true,
		},
		{
			name:     "minor takes precedence over patch",
			a:        "1.0.5",
			b:        "1.1.0",
			wantKind: "minor",
			wantNew:  true,
		},
		{
			name:    "invalid first version",
			a:       "bad",
			b:       "1.0.0",
			wantErr: true,
		},
		{
			name:    "invalid second version",
			a:       "1.0.0",
			b:       "bad",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, newer, err := CompareVersions(tt.a, tt.b)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, kind)
			assert.Equal(t, tt.wantNew, newer)
		})
	}
}
