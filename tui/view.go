package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	statusIndicator := "🟢"
	statusLabel := "CONNECTED"

	if m.Status == "Triggering graceful restart command..." {
		statusIndicator = "🟡"
		statusLabel = "RESTARTING..."
	} else if m.Err != nil || m.Status == "DISCONNECTED" {
		statusIndicator = "🔴"
		statusLabel = "DISCONNECTED"
	}

	// Format top console header bar dynamically
	headerText := fmt.Sprintf("  Internal Container Admin  |  %s", statusLabel)

	content := fmt.Sprintf(
		"\n  ┌──────────────────────────────────────────────────────────┐\n"+
			"  │ %-56s │\n"+
			"  └──────────────────────────────────────────────────────────┘\n\n"+
			"  [ SYSTEM METRICS ]\n"+
			"  • Connection Gateway :  %s\n"+
			"  • Engine Host OS     :  %s\n"+
			"  • Virtualized CPUs   :  %d cores\n"+
			"  • Go Engine Runtime  :  %s\n\n"+
			"  [ RUNTIME CONTAINER SCOPE ]\n"+
			"  • Container ID       :  %s\n"+
			"  • Container Name     :  %s\n"+
			"  • Active Image URI   :  %s\n\n"+
			"  [ STATUS ]\n"+
			"  • System State       :  %s %s\n",
		headerText,
		m.Term, m.Info.OS, m.Info.NumCPU, m.Info.GoVersion,
		m.Info.ID, m.Info.Name, m.Info.Image, statusIndicator, m.Status,
	)

	if m.Err != nil {
		content += fmt.Sprintf("  • Logged Error       :  %v\n", m.Err)
	}

	content += "\n  [ KEYBOARD COMMANDS ]\n" +
		"  • [r] Trigger Graceful Self-Restart\n" +
		"  • [q] Exit & Drop Terminal Connection\n"

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
