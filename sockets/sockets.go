//go:generate easyjson sockets.go

package sockets

import (
	"time"

	"golang.org/x/net/websocket"

	"github.com/dustinblackman/streamroller/logger"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

var (
	openSockets = map[string]*Socket{}
	// SocketChannel is the main channel where all chat messages are aggregated and passed to open web sockets
	SocketChannel = make(chan *SocketMessage, 10000)
)

// Socket holds information for a single socket
type Socket struct {
	ID       string
	SendChan chan []byte
}

// SocketMessage is used to submit a message entry back to websocket listeners.
// easyjson:json
type SocketMessage struct {
	Message string `json:"message"`
	Source  string `json:"source"`
	User    string `json:"user"`
}

// WriteToWebSocket is a quick helper function to write to an open web socket
func WriteToWebSocket(ws *websocket.Conn, message []byte) {
	_, err := ws.Write([]byte(message))
	if err != nil {
		logger.Log.Error(err)
	}
}

func listenSocketChannel() {
	for msg := range SocketChannel {
		data, err := msg.MarshalJSON()
		if err != nil {
			logger.Log.Error(err)
			continue
		}

		for _, socket := range openSockets {
			socket.SendChan <- data
		}
	}
}

// HandleWebSocketConnections sets up handling websocket connections for echo
func HandleWebSocketConnections() echo.HandlerFunc {
	go listenSocketChannel()

	return func(ctx echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				id := uuid.NewV4().String()
				sendChan := make(chan []byte, 100)
				openSockets[id] = &Socket{id, sendChan}
				go func() {
					for data := range sendChan {
						WriteToWebSocket(ws, data)
						time.Sleep(100 * time.Millisecond)
					}
				}()

				// Read
				msg := ""
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					break
				}
			}
		}).ServeHTTP(ctx.Response(), ctx.Request())

		return nil
	}
}
