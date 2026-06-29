package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/gorilla/websocket"
	"github.com/moby/moby/client"
	"github.com/peterjohnbishop/turbo-octo-potato/tui"
)

type wsWrapper struct {
	conn *websocket.Conn
	r    io.Reader
	mu   sync.Mutex // Mutex lock prevents multi-threaded WebSocket write crashes
}

func (w *wsWrapper) Read(p []byte) (int, error) {
	for {
		if w.r == nil {
			messageType, r, err := w.conn.NextReader()
			if err != nil {
				return 0, err
			}
			if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
				w.r = r
			} else {
				continue
			}
		}
		n, err := w.r.Read(p)
		if err == io.EOF {
			w.r = nil
			if n > 0 {
				return n, nil
			}
			continue
		}
		return n, err
	}
}

func (w *wsWrapper) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	cli, err := client.New(client.FromEnv)
	if err != nil {
		fmt.Printf("WebSocket Docker SDK initialization failed: %v\n", err)
	}

	width, height := 80, 24
	vp := viewport.New(width, height)

	m := tui.Model{
		Viewport:  vp,
		Term:      "xterm-web",
		Width:     width,
		Height:    height,
		DockerCli: cli,
		Status:    "Active Session Opened",
	}

	wrapper := &wsWrapper{conn: conn}
	p := tea.NewProgram(
		m,
		tea.WithInput(wrapper),
		tea.WithOutput(wrapper),
	)

	// Inject a concurrent WindowSizeMsg to trigger the layout inside Bubble Tea
	go func() {
		time.Sleep(50 * time.Millisecond)
		p.Send(tea.WindowSizeMsg{Width: width, Height: height})
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Runtime bubble tea canvas error: %v\n", err)
	}

	if cli != nil {
		cli.Close()
	}
}
