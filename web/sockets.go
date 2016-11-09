package web

import (
	"time"

	"golang.org/x/net/websocket"

	"github.com/Sirupsen/logrus"
	fastwebsocket "github.com/clevergo/websocket"
	"github.com/labstack/echo"
	fastengine "github.com/labstack/echo/engine/fasthttp"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var (
	socketUpgrader = fastwebsocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { // Allows all origins, limitations should be done by CORS.
			return true
		},
	}
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

func writeToWebSocket(ws *websocket.Conn, message string) {
	_, err := ws.Write([]byte(message))
	if err != nil {
		logrus.Error(err)
	}
}

func (s *EchoServer) sendMsgToListeners(msg *SocketMessage) {
	data, err := msg.MarshalJSON()
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, socket := range s.Sockets {
		socket.SendChan <- data
	}
}

// StartSocketConnection upgrades an echo GET request to a websocket connection
func (s *EchoServer) StartSocketConnection(ctx echo.Context) error {
	req := ctx.Request().(*fastengine.Request)
	err := socketUpgrader.Upgrade(req.RequestCtx, func(ws *fastwebsocket.Conn) {
		id := uuid.NewV4().String()
		sendChan := make(chan []byte, 100)
		s.Sockets[id] = &Socket{id, sendChan}
		go func() {
			for data := range sendChan {
				err := ws.WriteMessage(1, data)
				if err != nil {
					logrus.Error(err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()

		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}

		delete(s.Sockets, id)
	})

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
