package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/peterjohnbishop/turbo-octo-potato/server"
)

const (
	sshPort  = "2222"
	httpPort = "8080"
	host     = "0.0.0.0"
)

func main() {
	// Wish SSH Initialization
	s, err := server.StartSSH(host, sshPort)

	// GIN Routing Initialization
	r := server.ServeGin()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Starting SSH gateway on %s:%s\n", host, sshPort)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			fmt.Println("SSH Server error:", err)
			done <- nil
		}
	}()

	fmt.Printf("Starting Web gateway on http://%s:%s\n", host, httpPort)
	go func() {
		if err = http.ListenAndServe(net.JoinHostPort(host, httpPort), r); err != nil {
			fmt.Println("HTTP Server error:", err)
			done <- nil
		}
	}()

	<-done
	fmt.Println("\nStopping servers")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
