package tui

import tea "charm.land/bubbletea/v2"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ContainerInfo:
		m.cInfo = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.cInfo.Client != nil {
				return m, restartContainer(m.cInfo.Client, m.cInfo.ID)
			}
		}
	}
	return m, nil
}
