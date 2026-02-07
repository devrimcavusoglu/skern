package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestPrinter(jsonMode, quiet bool) (*Printer, *bytes.Buffer, *bytes.Buffer) {
	var outBuf, errBuf bytes.Buffer
	p := &Printer{
		out:    &outBuf,
		errOut: &errBuf,
		json:   jsonMode,
		quiet:  quiet,
	}
	return p, &outBuf, &errBuf
}

func TestPrinter_Print(t *testing.T) {
	tests := []struct {
		name   string
		json   bool
		quiet  bool
		format string
		args   []any
		want   string
	}{
		{
			name:   "text mode prints output",
			format: "hello %s",
			args:   []any{"world"},
			want:   "hello world",
		},
		{
			name:   "json mode suppresses Print",
			json:   true,
			format: "hello",
			want:   "",
		},
		{
			name:   "quiet mode suppresses Print",
			quiet:  true,
			format: "hello",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, out, _ := newTestPrinter(tt.json, tt.quiet)
			p.Print(tt.format, tt.args...)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestPrinter_PrintResult(t *testing.T) {
	tests := []struct {
		name string
		json bool
		quiet bool
		data any
		text string
		want string
	}{
		{
			name: "text mode uses text representation",
			data: map[string]string{"key": "value"},
			text: "key=value\n",
			want: "key=value\n",
		},
		{
			name: "json mode outputs JSON",
			json: true,
			data: map[string]string{"key": "value"},
			text: "key=value\n",
			want: "{\n  \"key\": \"value\"\n}\n",
		},
		{
			name:  "quiet mode suppresses all output",
			quiet: true,
			data:  map[string]string{"key": "value"},
			text:  "key=value\n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, out, _ := newTestPrinter(tt.json, tt.quiet)
			p.PrintResult(tt.data, tt.text)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestPrinter_PrintError(t *testing.T) {
	p, _, errOut := newTestPrinter(false, false)
	p.PrintError("something went %s", "wrong")
	assert.Equal(t, "Error: something went wrong\n", errOut.String())
}

func TestPrinter_PrintError_QuietStillShows(t *testing.T) {
	p, _, errOut := newTestPrinter(false, true)
	p.PrintError("still visible")
	assert.Equal(t, "Error: still visible\n", errOut.String())
}

func TestPrinter_PrintError_JSONSuppresses(t *testing.T) {
	p, _, errOut := newTestPrinter(true, false)
	p.PrintError("hidden in json mode")
	assert.Equal(t, "", errOut.String())
}

func TestPrinter_PrintErrorResult_JSON(t *testing.T) {
	p, out, _ := newTestPrinter(true, false)
	p.PrintErrorResult(assert.AnError)
	assert.Contains(t, out.String(), `"error"`)
}

func TestPrinter_Flags(t *testing.T) {
	p := NewPrinter(true, true)
	assert.True(t, p.IsJSON())
	assert.True(t, p.IsQuiet())

	p2 := NewPrinter(false, false)
	assert.False(t, p2.IsJSON())
	assert.False(t, p2.IsQuiet())
}

func TestVersionResult_JSON(t *testing.T) {
	p, out, _ := newTestPrinter(true, false)
	p.PrintResult(VersionResult{
		Version: "0.0.1",
		Commit:  "abc1234",
		Date:    "2026-02-07",
	}, "")
	assert.Contains(t, out.String(), `"version": "0.0.1"`)
	assert.Contains(t, out.String(), `"commit": "abc1234"`)
}
