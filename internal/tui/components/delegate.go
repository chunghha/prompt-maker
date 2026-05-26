package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ModelOption defines the interface for a selectable model in the TUI list.
type ModelOption interface {
	Name() string
	Desc() string
}

// ItemDelegate for the model selection list.
type ItemDelegate struct{}

// Height returns the height of a single list item.
func (ItemDelegate) Height() int { return 1 }

// Spacing returns the spacing between list items.
func (ItemDelegate) Spacing() int { return 0 }

// Update handles messages for the delegate (no-op).
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render draws a single list item to the writer.
//
//nolint:gocritic // The signature is defined by the list.ItemDelegate interface, which requires passing list.Model by value.
func (ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ModelOption)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s (%s)", index+1, i.Name(), i.Desc())
	styles := NewStyles() // Create a new Styles struct to access the styles.

	var fn func(...string) string
	if index == m.Index() {
		fn = func(s ...string) string {
			return styles.SelectedListItem.Render("> " + strings.Join(s, " "))
		}
	} else {
		fn = func(s ...string) string {
			return styles.ListItem.Render(strings.Join(s, " "))
		}
	}

	_, _ = io.WriteString(w, fn(str))
}
