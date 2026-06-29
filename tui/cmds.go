package tui

import (
	"context"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/moby/moby/client"
)

func fetchContainerInfo(cli *client.Client) tea.Cmd {
	return func() tea.Msg {
		info := ContainerInfo{
			OS:        runtime.GOOS,
			NumCPU:    runtime.NumCPU(),
			GoVersion: runtime.Version(),
		}

		hostname, err := os.Hostname()
		if err == nil {
			info.ID = hostname
		} else {
			info.ID = "Unknown"
		}

		if cli != nil {
			inspect, err := cli.ContainerInspect(context.Background(), info.ID, client.ContainerInspectOptions{})
			if err == nil {
				info.Name = inspect.Container.Name
				info.Image = inspect.Container.Config.Image
			} else {
				info.Name = "Unknown (Missing Socket Mount)"
				info.Image = "Unknown"
			}
		} else {
			info.Name = "Client Connection Uninitialized"
			info.Image = "Unknown"
		}

		return info
	}
}

func restartContainer(cli *client.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		// can pass a custom timeout value
		options := client.ContainerRestartOptions{}

		_, err := cli.ContainerRestart(context.Background(), containerID, options)
		if err != nil {
			return errMsg{err} // Return an error message to handle in your Update loop
		}

		return successMsg("Container restarted successfully")
	}
}
