package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	bm "charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/ssh"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/peterjohnbishop/turbo-octo-potato/tui"
)

const (
	sshPort  = "2222"
	httpPort = "8080"
	host     = "0.0.0.0"
)

const htmlClient = `<!DOCTYPE html>
<html>
  <head>
    <title>Turbo-Octo-Potato Web Gateway</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-attach@0.9.0/lib/xterm-addon-attach.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js"></script>
    <style>
      body { background: #000; margin: 0; padding: 20px; height: 100vh; box-sizing: border-box; }
      #terminal { width: 100%; height: 100%; }
    </style>
  </head>
  <body>
    <div id="terminal"></div>
    <script>
      // FIX: Added convertEol: true to prevent the staircase/cascading text effect
      const term = new Terminal({ 
        cursorBlink: true, 
        convertEol: true, 
        theme: { background: '#000000' } 
      });
      
      const fitAddon = new FitAddon.FitAddon();
      term.loadAddon(fitAddon);
      term.open(document.getElementById('terminal'));
      fitAddon.fit();

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
      
      const attachAddon = new AttachAddon.AttachAddon(ws);
      term.loadAddon(attachAddon);

      window.addEventListener('resize', () => fitAddon.fit());
    </script>
  </body>
</html>`

// Upgrader configuration for Gorilla Websockets
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust security settings for production
	},
}

type wsWrapper struct {
	conn *websocket.Conn
	r    io.Reader
}

func (w *wsWrapper) Read(p []byte) (int, error) {
	if w.r == nil {
		_, r, err := w.conn.NextReader()
		if err != nil {
			return 0, err
		}
		w.r = r
	}
	n, err := w.r.Read(p)
	if err == io.EOF {
		w.r = nil
		return n, nil
	}
	return n, err
}

func (w *wsWrapper) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func main() {
	// Wish
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
		fmt.Println("Could not start SSH server:", err)
		os.Exit(1)
	}

	// GIN
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlClient))
	})

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

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
		if err = http.ListenAndServe(net.JoinHostPort(host, httpPort), router); err != nil {
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

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		fmt.Println("No active terminal, skipping")
		return nil, nil
	}

	vp := viewport.New(pty.Window.Width, pty.Window.Height)

	m := tui.Model{
		Viewport: vp,
		Term:     pty.Term,
		Width:    pty.Window.Width,
		Height:   pty.Window.Height,
	}

	return m, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Websocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	rw := &wsWrapper{conn: conn}

	width, height := 80, 24
	vp := viewport.New(width, height)

	m := tui.Model{
		Viewport: vp,
		Term:     "xterm-web",
		Width:    width,
		Height:   height,
	}

	p := tea.NewProgram(
		m,
		tea.WithInput(rw),
		tea.WithOutput(rw),
	)

	go func() {
		time.Sleep(50 * time.Millisecond)
		p.Send(tea.WindowSizeMsg{Width: width, Height: height})
	}()

	_, _ = p.Run()
}
