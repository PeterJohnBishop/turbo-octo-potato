package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	content := fmt.Sprintf(
		"\n  Welcome to my Docker Container\n\n  Terminal: %s\n  Window Size: %dx%d\n\n  Press 'q' or 'ctrl+c' to disconnect.",
		m.Term, m.Width, m.Height,
	)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
