package server

import (
	"fmt"
	"net"

	tea "charm.land/bubbletea/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	bm "charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	"github.com/moby/moby/client"
	"github.com/peterjohnbishop/turbo-octo-potato/tui"
)

func StartSSH(host, sshPort string) (*ssh.Server, error) {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, sshPort)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			logging.Middleware(),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		return nil, err
		//fmt.Println("Could not start SSH server:", err)
		//os.Exit(1)
	}
	return s, nil
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		return nil, nil
	}

	cli, err := client.New(client.FromEnv)
	if err != nil {
		fmt.Printf("SSH Docker SDK initialization failed: %v\n", err)
	}

	m := tui.Model{
		Term:      "wish/ssh",
		Width:     pty.Window.Width,
		Height:    pty.Window.Height,
		DockerCli: cli,
		Status:    "Active Session Opened",
	}
	return m, nil
}
