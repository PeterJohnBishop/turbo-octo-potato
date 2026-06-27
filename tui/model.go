package tui

import (
	"github.com/charmbracelet/bubbles/viewport"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Viewport   viewport.Model
	Term       string
	Width      int
	Height     int
	cInfo      ContainerInfo
	err        error
	ClientType string
}

func (m Model) Init() tea.Cmd {
	return fetchContainerInfo()
}
