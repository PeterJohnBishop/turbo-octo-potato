package tui

import tea "charm.land/bubbletea/v2"

type Model struct {
	Term   string
	Width  int
	Height int
}

func (m Model) Init() tea.Cmd {
	return nil
}
