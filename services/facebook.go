package services

import (
	"errors"
	"strings"
	"time"

	"github.com/laice/joy4/format/rtmp"
	"github.com/laice/streamroller/logger"
	"github.com/laice/streamroller/sockets"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// FacebookService handles RTMP and chat for Facebook
type FacebookService struct {
	name string
}

// Name returns the name of the service
func (f *FacebookService) Name() string {
	return f.name
}

// CanConnect returns a bool whether all configuration is available to connect to RTMP
func (f *FacebookService) CanConnect() bool {
	return viper.GetViper().GetString("facebook-livekey") != ""
}

// ConnectRTMP connects to RTMP server
func (f *FacebookService) ConnectRTMP() *rtmp.Conn {
	logger.Log.Info("Connecting to Facebook RTMP")
	return connectRTMP("rtmp://rtmp-api.facebook.com:80/rtmp/" + viper.GetViper().GetString("facebook-livekey"))
}

func (f *FacebookService) connectChat(videoID, accessToken string) {
	logger.Log.Info("Starting facebook chat monitor")
	var lastID string
	sleepDuration := 3000 * time.Millisecond
	for {
		// TODO Handle error codes
		_, body, _ := gorequest.New().
			Get("https://graph.facebook.com/v2.8/" + videoID + "/comments").
			Query(`{"order": "reverse_chronological", "access_token": "` + accessToken + `"}`).
			End()

		newMessages := []sockets.SocketMessage{}
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

			newMessages = append(newMessages, sockets.SocketMessage{
				Message: message,
				Source:  f.name,
				User:    name,
			})
		}

		lastID = string(nextLastID)
		for _, socketMessage := range newMessages {
			sockets.SocketChannel <- &socketMessage
		}
		time.Sleep(sleepDuration)
	}
}

// ConnectChat connects to chat service if available
func (f *FacebookService) ConnectChat() error {
	viper := viper.GetViper()
	if viper.GetString("facebook-livekey") != "" && viper.GetString("facebook-token") != "" {
		videoID := strings.Split(viper.GetString("facebook-livekey"), "?")[0]
		go f.connectChat(videoID, viper.GetString("facebook-token"))
		return nil
	}

	return errors.New("Missing keys")
}

func init() {
	RegisterService(&FacebookService{name: "facebook"})
}
