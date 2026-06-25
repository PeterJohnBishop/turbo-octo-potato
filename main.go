package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	bm "charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	"github.com/peterjohnbishop/turbo-octo-potato/tui"
)

const (
	host = "0.0.0.0" // Listen on all interfaces for Docker
	port = "2222"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			logging.Middleware(),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		fmt.Println("Could not start server:", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Starting SSH server on %s:%s\n", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			fmt.Println("Could not start server:", err)
			done <- nil
		}
	}()

	<-done
	fmt.Println("\nStopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		fmt.Println("Could not stop server:", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		fmt.Println("No active terminal, skipping")
		return nil, nil
	}

	m := tui.Model{
		Term:   pty.Term,
		Width:  pty.Window.Width,
		Height: pty.Window.Height,
	}

	// In v2, options like AltScreen are now set declaratively on the View.
	return m, nil
}
