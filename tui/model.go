package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/moby/moby/client"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Viewport  viewport.Model
	Term      string
	Width     int
	Height    int
	Info      ContainerInfo
	DockerCli *client.Client
	Status    string
	Err       error
}

func (m Model) Init() tea.Cmd {
	return fetchContainerInfo(m.DockerCli)
}
