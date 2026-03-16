package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LoadOmarchyTheme attempts to read the active Omarchy theme from
// ~/.config/omarchy/themes/current/colors.toml and maps it to a Theme.
// Returns (theme, true) on success, (zero, false) if Omarchy isn't present.
func LoadOmarchyTheme() (Theme, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Theme{}, false
	}

	themeDir := filepath.Join(home, ".config", "omarchy", "current", "theme")
	data, err := os.ReadFile(filepath.Join(themeDir, "colors.toml"))
	if err != nil {
		return Theme{}, false
	}

	c := parseOmarchyColors(data)
	if c["foreground"] == "" || c["color1"] == "" {
		return Theme{}, false // incomplete / unrecognised format
	}

	// Light mode is signalled by an empty light.mode file in the theme dir.
	_, isLight := os.Stat(filepath.Join(themeDir, "light.mode"))
	light := isLight == nil

	// adaptive returns an AdaptiveColor using the same hex for both modes.
	// Lipgloss picks the right variant based on terminal background, which
	// Omarchy sets to match — so both sides being identical is correct.
	adaptive := func(hex string) lipgloss.AdaptiveColor {
		return lipgloss.AdaptiveColor{Light: hex, Dark: hex}
	}

	// For Selected background, invert based on light/dark so the selection
	// colour doesn't get buried.
	selectedBg := c["selection_background"]
	if selectedBg == "" {
		if light {
			selectedBg = c["color4"] // normal blue
		} else {
			selectedBg = c["color4"]
		}
	}

	// Border: use bright-black (color8) — muted but visible.
	border := c["color8"]
	if border == "" {
		border = c["color7"]
	}

	return Theme{
		Online:        lipgloss.Color(c["color2"]),  // normal green
		Idle:          lipgloss.Color(c["color3"]),  // normal yellow
		Offline:       lipgloss.Color(c["color1"]),  // normal red
		Unknown:       adaptive(c["color8"]),         // bright black
		Accent:        lipgloss.Color(c["accent"]),
		AccentSubtle:  adaptive(c["color5"]),         // normal magenta/purple
		Selected:      adaptive(selectedBg),
		Border:        adaptive(border),
		TextPrimary:   adaptive(c["foreground"]),
		TextSecondary: adaptive(c["color8"]),         // bright black / comment
	}, true
}

// parseOmarchyColors parses the simple key = "#hex" format used by
// colors.toml without pulling in a TOML library.
func parseOmarchyColors(data []byte) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if key != "" && val != "" {
			result[key] = val
		}
	}
	return result
}
