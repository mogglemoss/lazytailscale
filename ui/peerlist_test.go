package ui

import (
	"testing"
	"unicode/utf8"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},        // shorter than max — unchanged
		{"hello", 5, "hello"},         // exact length — unchanged
		{"hello", 4, "hel…"},          // truncated: 3 chars + ellipsis = 4 runes
		{"hello", 1, "…"},             // max=1 → ellipsis only
		{"hello", 0, "…"},             // max=0 → ellipsis (degenerate)
		{"", 5, ""},                   // empty string
		{"日本語テスト", 4, "日本語…"}, // multibyte runes
	}

	for _, tc := range tests {
		got := truncate(tc.input, tc.max)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
		// Result must never exceed max rune count.
		if utf8.RuneCountInString(got) > tc.max && tc.max > 0 {
			t.Errorf("truncate(%q, %d) result %q exceeds max rune count %d",
				tc.input, tc.max, got, tc.max)
		}
	}
}

func TestOsTag(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"linux", "Linux"},
		{"Linux", "Linux"},
		{"LINUX", "Linux"},
		{"darwin", "Mac"},
		{"macos", "Mac"},
		{"windows", "Win"},
		{"android", "Android"},
		{"ios", "iOS"},
		{"freebsd", "freebsd"}, // unknown — pass through
		{"", ""},               // empty — empty
	}

	for _, tc := range tests {
		got := osTag(tc.input)
		if got != tc.want {
			t.Errorf("osTag(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
