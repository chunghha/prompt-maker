package web

const (
	// DefaultTheme is the default theme for the web interface.
	DefaultTheme = "prompt-maker"
)

// Theme represents a single theme with its display label and group.
type Theme struct {
	ID    string
	Label string
	Group string // "light" or "dark"
}

// getThemes returns the list of all available themes for the dropdown.
func getThemes() []Theme {
	return []Theme{
		// Light themes
		{ID: "prompt-maker", Label: "Pastel Sketchbook", Group: "light"},
		{ID: "default-light", Label: "Default", Group: "light"},
		{ID: "gruvbox-light", Label: "Gruvbox", Group: "light"},
		{ID: "solarized-light", Label: "Solarized", Group: "light"},
		{ID: "flexoki-light", Label: "Flexoki", Group: "light"},
		{ID: "ayu-light", Label: "Ayu", Group: "light"},
		{ID: "zoegi-light", Label: "Zoegi", Group: "light"},
		{ID: "ffe-light", Label: "FFE", Group: "light"},
		{ID: "postrboard-light", Label: "Postrboard", Group: "light"},
		// Dark themes
		{ID: "prompt-maker-dark", Label: "Pastel Sketchbook", Group: "dark"},
		{ID: "default-dark", Label: "Default", Group: "dark"},
		{ID: "gruvbox", Label: "Gruvbox", Group: "dark"},
		{ID: "solarized", Label: "Solarized", Group: "dark"},
		{ID: "ayu", Label: "Ayu", Group: "dark"},
		{ID: "flexoki", Label: "Flexoki", Group: "dark"},
		{ID: "zoegi", Label: "Zoegi", Group: "dark"},
		{ID: "ffe-dark", Label: "FFE", Group: "dark"},
		{ID: "postrboard", Label: "Postrboard", Group: "dark"},
	}
}
