package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {

	var content string

	if m.ClientType == "ssh" {
		content = fmt.Sprintf(
			"\n  [ Connection: SSH ]\n\n"+
				"  Container ID:    %s\n"+
				"  Container Name:  %s\n"+
				"  Image Signature: %s\n"+
				"  Engine Host OS:  %s\n"+
				"  Allocated CPUs:  %d\n"+
				"  Go Version:      %s\n\n"+
				"  [Press 'q' or 'ctrl+c' to close the terminal session]",
			m.cInfo.ID, m.cInfo.Name, m.cInfo.Image, m.cInfo.OS, m.cInfo.NumCPU, m.cInfo.GoVersion,
		)
	} else {
		content = fmt.Sprintf(
			"\n  [ Connection: Web ]\n\n"+
				"  Container ID:    %s\n"+
				"  Container Name:  %s\n"+
				"  Image Signature: %s\n"+
				"  Engine Host OS:  %s\n"+
				"  Allocated CPUs:  %d\n"+
				"  Go Version:      %s\n\n"+
				"  [Press 'q' or 'ctrl+c' to close the terminal session]",
			m.cInfo.ID, m.cInfo.Name, m.cInfo.Image, m.cInfo.OS, m.cInfo.NumCPU, m.cInfo.GoVersion,
		)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
