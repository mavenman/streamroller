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
	Sockets map[string]*Socket
}

func refreshYoutubeToken(refreshToken string) (string, int64) {
	// TODO Handle errors
	_, body, _ := gorequest.New().
		Post("https://developers.google.com/oauthplayground/refreshAccessToken").
		Type("json").
		Send(`{"refresh_token":"` + refreshToken + `","token_uri":"https://www.googleapis.com/oauth2/v4/token"}`).
		End()

	accessToken := gjson.Get(body, "access_token").String()
	expiresIn := gjson.Get(body, "expires_in").Int()

	return accessToken, expiresIn
}

func getYoutubeChatID(accessToken string) string {
	_, body, _ := gorequest.New().
		Get("https://www.googleapis.com/youtube/v3/liveBroadcasts?part=snippet&broadcastStatus=active&broadcastType=all").
		Set("Authorization", "Bearer "+accessToken).
		End()

	return gjson.Get(body, "items.0.snippet.liveChatId").String()
}

// ListenYoutubeChat monitors for new streams to start and listens for chat messages if users are listening on websockets.
func (s *EchoServer) ListenYoutubeChat(refreshToken string) {
	log := logrus.WithFields(logrus.Fields{"module": "web", "method": "youtube"})
	log.Info("Starting youtube chat monitor")
	accessToken, expiresIn := refreshYoutubeToken(refreshToken)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(expiresIn-100))
			accessToken, expiresIn = refreshYoutubeToken(refreshToken)
		}
	}()

	liveChatID := getYoutubeChatID(accessToken)
	go func() {
		for {
			time.Sleep(time.Second * 15)
			liveChatID = getYoutubeChatID(accessToken)
		}
	}()

	var nextPageToken string
	sleepDuration := 3000 * time.Millisecond
	first := true
	for {
		if liveChatID == "" {
			time.Sleep(sleepDuration)
			continue
		}
		log.Debug("Polling Youtube")

		url := "https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet,authorDetails&liveChatId=" + liveChatID
		if nextPageToken != "" {
			url += "&pageToken=" + nextPageToken
		}
		_, body, _ := gorequest.New().
			Get(url).
			Set("Authorization", "Bearer "+accessToken).
			End()

		newMessages := []SocketMessage{}
		nextPageToken = gjson.Get(body, "nextPageToken").String()
		sleepDuration = time.Duration(gjson.Get(body, "pollingIntervalMillis").Int()) * time.Millisecond

		// Ignore first request to prevent flooding with history
		if first {
			first = false
			time.Sleep(sleepDuration)
			continue
		}

		for _, entry := range gjson.Get(body, "items").Array() {
			name := gjson.Get(entry.Raw, "authorDetails.displayName").String()
			message := gjson.Get(entry.Raw, "snippet.textMessageDetails.messageText").String()

			log.Debug(spew.Sdump(SocketMessage{message, "youtube", name}))
			newMessages = append(newMessages, SocketMessage{message, "youtube", name})
		}

		for _, socketMessage := range newMessages {
			s.sendMsgToListeners(&socketMessage)
		}
		time.Sleep(sleepDuration)
	}
}

// ListenFacebookChat starts polling Facebook's API every 3 seconds for chat messages.
func (s *EchoServer) ListenFacebookChat(videoID, accessToken string) {
	log := logrus.WithFields(logrus.Fields{"module": "web", "method": "facebook"})
	log.Info("Starting facebook chat monitor")

	var lastID string
	sleepDuration := 3000 * time.Millisecond
	for {
		log.Debug("Polling Facebook")
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
			if lastID == id {
				break
			}
			log.Debug(spew.Sdump(SocketMessage{message, "facebook", name}))
			newMessages = append(newMessages, SocketMessage{message, "facebook", name})
		}

		lastID = string(nextLastID)
		for _, socketMessage := range newMessages {
			s.sendMsgToListeners(&socketMessage)
		}
		time.Sleep(sleepDuration)
	}
}

// ListenTwitchChat logs in to a Twitch channels IRC and submits messages back websocket listeners
func (s *EchoServer) ListenTwitchChat(username, oauthKey string) {
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
		go server.ListenTwitchChat(viper.GetString("twitch-username"), viper.GetString("twitch-oauth"))
	}
	if viper.GetString("facebook-livekey") != "" && viper.GetString("facebook-token") != "" {
		videoID := strings.Split(viper.GetString("facebook-livekey"), "?")[0]
		go server.ListenFacebookChat(videoID, viper.GetString("facebook-token"))
	}
	if viper.GetString("youtube-livekey") != "" && viper.GetString("youtube-token") != "" {
		go server.ListenYoutubeChat(viper.GetString("youtube-token"))
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
