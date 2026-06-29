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
	"sync"
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
	"github.com/moby/moby/client"
	"github.com/peterjohnbishop/turbo-octo-potato/tui"
)

const (
	sshPort  = "2222"
	httpPort = "8080"
	host     = "0.0.0.0"
)

// Embedded premium web frontend client
const htmlClient = `<!DOCTYPE html>
<html>
  <head>
    <title>Internal Container Admin</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&display=swap" rel="stylesheet">
    <style>
      body {
        background: #000000;
        margin: 0;
        padding: 0;
        height: 100vh;
        width: 100vw;
        display: flex;
        overflow: hidden;
      }
      .terminal-container {
        flex: 1;
        width: 100%;
        height: 100%;
        background-color: #000000;
      }
      #terminal {
        width: 100%;
        height: 100%;
        padding: 16px;
        box-sizing: border-box;
        background-color: #000000;
      }
    </style>
  </head>
  <body>
    <div class="terminal-container">
      <div id="terminal"></div>
    </div>
    <script>
      const term = new Terminal({
        cursorBlink: true,
        convertEol: true, // Fixes the staircase effect
        fontFamily: 'JetBrains Mono, Courier New, monospace',
        fontSize: 14,
        theme: {
          background: '#000000',
          foreground: '#f3f4f6',
          cursor: '#818cf8',
        }
      });
      const fitAddon = new FitAddon.FitAddon();
      term.loadAddon(fitAddon);
      term.open(document.getElementById('terminal'));
      fitAddon.fit();

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = protocol + '//' + window.location.host + '/ws';
      
      let ws;
      let reconnectDelay = 1000;
      const maxReconnectDelay = 5000;

      function connect() {
        ws = new WebSocket(wsUrl);
        ws.binaryType = 'arraybuffer';

        ws.onopen = () => {
          reconnectDelay = 1000;
          term.clear();
        };

        ws.onclose = () => {
          setTimeout(() => {
            connect();
          }, reconnectDelay);
          reconnectDelay = Math.min(reconnectDelay * 1.5, maxReconnectDelay);
        };

        ws.onerror = (err) => {
          console.error("WebSocket session errored out: ", err);
          ws.close();
        };

        ws.onmessage = (event) => {
          term.write(new Uint8Array(event.data));
        };
      }

      term.onData(data => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(data);
        }
      });

      window.addEventListener('resize', () => {
        fitAddon.fit();
      });

      connect();
    </script>
  </body>
</html>`

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

func main() {
	// Wish SSH Initialization
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

	// GIN Routing Initialization
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
		return nil, nil
	}

	cli, err := client.New(client.FromEnv)
	if err != nil {
		fmt.Printf("SSH Docker SDK initialization failed: %v\n", err)
	}

	m := tui.Model{
		Term:      pty.Term,
		Width:     pty.Window.Width,
		Height:    pty.Window.Height,
		DockerCli: cli,
		Status:    "Active Session Opened",
	}
	return m, nil
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
