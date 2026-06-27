package tui

import (
	"context"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/moby/moby/client"
)

func fetchContainerInfo() tea.Cmd {
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

		cli, err := client.New(client.FromEnv)
		if err == nil {
			defer cli.Close()
			inspect, err := cli.ContainerInspect(context.Background(), info.ID, client.ContainerInspectOptions{})
			if err == nil {
				info.Name = inspect.Container.Name
				info.Image = inspect.Container.Config.Image
			} else {
				info.Name = "Unknown (Requires Docker Socket mount)"
				info.Image = "Unknown"
			}
		}

		return info
	}
}
