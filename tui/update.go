package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ContainerInfo:
		m.Info = msg
		m.Status = "CONNECTED"
		return m, nil

	case successMsg:
		m.Status = string(msg)
		return m, nil

	case errMsg:
		m.Err = msg.err
		m.Status = "DISCONNECTED"
		return m, nil

	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.Status = "Triggering graceful restart command..."
			return m, restartContainer(m.DockerCli, m.Info.ID)
		}
	}
	return m, nil
}
