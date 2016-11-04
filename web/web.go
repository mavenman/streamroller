package web

import (
	"mime"
	"path"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo"
	fastengine "github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// EchoServer holds server information such as active sockets
type EchoServer struct {
	LastFacebookID string
	Sockets        map[string]*Socket
}

// StartFacebookChat starts polling Facebook's API every 3 seconds for chat messages if users are listening on websockets
func (s *EchoServer) StartFacebookChat(videoID, accessToken string) {
	log := logrus.WithFields(logrus.Fields{"module": "web", "method": "facebook"})
	log.Info("Starting facebook chat monitor")

	for {
		if len(s.Sockets) == 0 {
			time.Sleep(3000 * time.Millisecond)
			continue
		}

		log.Debug("Polling")
		// TODO Handle error codes
		_, body, _ := gorequest.New().
			Get("https://graph.facebook.com/v2.8/" + videoID + "/comments").
			Query(`{"order": "reverse_chronological", "access_token": "` + accessToken + `"}`).
			End()

		newMessages := []SocketMessage{}
		var nextLastID string
		for idx, entry := range gjson.Get(body, "data").Array() {
			name := gjson.Get(entry.Raw, "from.name").String()
			message := gjson.Get(entry.Raw, "message").String()
			createdAt := gjson.Get(entry.Raw, "created_time").String()
			id := name + message + createdAt

			if idx == 0 {
				nextLastID = string(id)
			}
			if s.LastFacebookID == id {
				break
			}
			log.Debug(spew.Sdump(SocketMessage{message, "facebook", name}))
			newMessages = append(newMessages, SocketMessage{message, "facebook", name})
		}

		s.LastFacebookID = string(nextLastID)
		for _, socketMessage := range newMessages {
			s.sendMsgToListeners(&socketMessage)
		}
		time.Sleep(3000 * time.Millisecond)
	}
}

// StartTwitchChat logs in to a Twitch channels IRC and submits messages back websocket listeners
func (s *EchoServer) StartTwitchChat(username, oauthKey string) {
	log := logrus.WithFields(logrus.Fields{"module": "web", "method": "twitch"})
	log.Info("Starting twitch chat monitor")
	ws, err := websocket.Dial("wss://irc-ws.chat.twitch.tv", "", "http://localhost")
	if err != nil {
		log.Error(err)
	}

	writeToWebSocket(ws, "CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership")
	writeToWebSocket(ws, "PASS oauth:"+strings.Replace(oauthKey, "oauth:", "", -1))
	writeToWebSocket(ws, "NICK "+username)
	writeToWebSocket(ws, "JOIN #"+username)

	for {
		msgByte := make([]byte, 2048)
		_, err = ws.Read(msgByte)
		if err != nil {
			log.Fatal(err)
		}
		msg := string(msgByte[:])
		if strings.Contains(msg, "PING :tmi.twitch.tv") {
			writeToWebSocket(ws, "PONG :tmi.twitch.tv")
		}

		if strings.Contains(msg, "PRIVMSG") {
			userMsg := strings.Split(strings.Split(msg, "PRIVMSG #"+username+" :")[1], "\r\n")[0]
			headers := map[string]string{}
			for _, entry := range strings.Split(msg, ";") {
				keyVal := strings.Split(entry, "=")
				headers[keyVal[0]] = strings.Join(keyVal[1:], "=")
			}
			s.sendMsgToListeners(&SocketMessage{userMsg, "twitch", headers["display-name"]})
		}
	}
}

// CreateEcho creates the HTTP server.
func CreateEcho(port string) {
	log := logrus.WithFields(logrus.Fields{"module": "web"})
	viper := viper.GetViper()
	server := EchoServer{}
	server.Sockets = map[string]*Socket{}

	if viper.GetString("twitch-username") != "" && viper.GetString("twitch-oauth") != "" {
		go server.StartTwitchChat(viper.GetString("twitch-username"), viper.GetString("twitch-oauth"))
	}
	if viper.GetString("facebook-livekey") != "" && viper.GetString("facebook-token") != "" {
		videoID := strings.Split(viper.GetString("facebook-livekey"), "?")[0]
		go server.StartFacebookChat(videoID, viper.GetString("facebook-token"))
	}

	app := echo.New()
	app.Use(middleware.Gzip())
	app.Use(middleware.CORS())
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// go-bindata middleware
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			uri := ctx.Request().URI()
			data, err := Asset("static" + uri)
			if err != nil {
				return next(ctx)
			}
			return ctx.Blob(200, mime.TypeByExtension(path.Ext(uri)), data)
		}
	})

	app.Get("/", func(ctx echo.Context) error {
		data, err := Asset("static/index.html")
		if err != nil {
			return ctx.String(404, "Not found")
		}
		return ctx.HTML(200, string(data))
	})
	app.Get("/socket", server.StartSocketConnection)

	log.Info("Starting Echo")
	go app.Run(fastengine.New(":" + port))
}
