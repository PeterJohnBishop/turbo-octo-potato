// Package server implements SSH server, Gin routing, and websocket handlers
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func ServeGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlClient))
	})

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

	return router
}
