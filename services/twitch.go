package services

import (
	"errors"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/dustinblackman/joy4/format/rtmp"
	"github.com/dustinblackman/streamroller/logger"
	"github.com/dustinblackman/streamroller/sockets"
	"github.com/spf13/viper"
)

// TwitchService handles RTMP and chat for Twitch
type TwitchService struct {
	name string
}

// Name returns the name of the service
func (t *TwitchService) Name() string {
	return t.name
}

// CanConnect returns a bool whether all configuration is available to connect to RTMP
func (t *TwitchService) CanConnect() bool {
	return viper.GetViper().GetString("twitch-livekey") != ""
}

// ConnectRTMP connects to RTMP server
func (t *TwitchService) ConnectRTMP() *rtmp.Conn {
	logger.Log.Info("Connecting to Twitch RTMP")
	return connectRTMP("rtmp://live.twitch.tv/app/" + viper.GetViper().GetString("twitch-livekey"))
}

func (t *TwitchService) connectChat(username, oauthKey string) {
	logger.Log.Info("Starting twitch chat monitor")
	ws, err := websocket.Dial("wss://irc-ws.chat.twitch.tv", "", "http://localhost")
	if err != nil {
		logger.Log.Error(err)
	}

	sockets.WriteToWebSocket(ws, []byte("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"))
	sockets.WriteToWebSocket(ws, []byte("PASS oauth:"+strings.Replace(oauthKey, "oauth:", "", -1)))
	sockets.WriteToWebSocket(ws, []byte("NICK "+username))
	sockets.WriteToWebSocket(ws, []byte("JOIN #"+username))

	for {
		msgByte := make([]byte, 2048)
		_, err = ws.Read(msgByte)
		if err != nil {
			// TODO: Should not be fatal
			logger.Log.Fatal(err)
		}
		msg := string(msgByte[:])
		if strings.Contains(msg, "PING :tmi.twitch.tv") {
			sockets.WriteToWebSocket(ws, []byte("PONG :tmi.twitch.tv"))
		}

		if strings.Contains(msg, "PRIVMSG") {
			userMsg := strings.Split(strings.Split(msg, "PRIVMSG #"+username+" :")[1], "\r\n")[0]
			headers := map[string]string{}
			for _, entry := range strings.Split(msg, ";") {
				keyVal := strings.Split(entry, "=")
				headers[keyVal[0]] = strings.Join(keyVal[1:], "=")
			}
			sockets.SocketChannel <- &sockets.SocketMessage{userMsg, t.name, headers["display-name"]}
		}
	}
}

// ConnectChat connects to chat service if available
func (t *TwitchService) ConnectChat() error {
	viper := viper.GetViper()
	if viper.GetString("twitch-username") != "" && viper.GetString("twitch-oauth") != "" {
		go t.connectChat(viper.GetString("twitch-username"), viper.GetString("twitch-oauth"))
		return nil
	}
	return errors.New("Missing keys")
}

func init() {
	RegisterService(&TwitchService{name: "twitch"})
}
