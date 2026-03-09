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

func (ItemDelegate) Height() int                             { return 1 }
func (ItemDelegate) Spacing() int                            { return 0 }
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

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
