package tui

import "github.com/moby/moby/client"

type ContainerInfo struct {
	Client    *client.Client
	ID        string
	Name      string
	Image     string
	OS        string
	NumCPU    int
	GoVersion string
}

type errMsg struct {
	err error
}

type successMsg string
