package web

const (
	// DefaultTheme is the default theme for the web interface.
	DefaultTheme = "milkshake"
)

// getThemes returns the list of all available DaisyUI themes for the dropdown.
func getThemes() []string {
	return []string{
		"light", "dark", "cupcake", "bumblebee", "emerald", "corporate",
		"synthwave", "retro", "cyberpunk", "valentine", "halloween", "garden",
		"forest", "aqua", "lofi", "pastel", "fantasy", "wireframe", "black",
		"luxury", "dracula", "cmyk", "autumn", "business", "acid", "lemonade",
		"night", "coffee", "winter", "dim", "nord", "sunset", "milkshake",
		"mindful", "pursuit",
	}
}
