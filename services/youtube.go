package services

import (
	"errors"
	"time"

	"github.com/laice/joy4/format/rtmp"
	"github.com/laice/streamroller/logger"
	"github.com/laice/streamroller/sockets"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// YoutubeService handles RTMP and chat for Youtube
type YoutubeService struct {
	name string
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

// Name returns the name of the service
func (y *YoutubeService) Name() string {
	return y.name
}

// CanConnect returns a bool whether all configuration is available to connect to RTMP
func (y *YoutubeService) CanConnect() bool {
	return viper.GetViper().GetString("youtube-livekey") != ""
}

// ConnectRTMP connects to RTMP server
func (y *YoutubeService) ConnectRTMP() *rtmp.Conn {
	logger.Log.Info("Connecting to Youtube RTMP")
	return connectRTMP("rtmp://a.rtmp.youtube.com/live2/" + viper.GetViper().GetString("youtube-livekey"))
}

func (y *YoutubeService) connectChat(refreshToken string) {
	logger.Log.Info("Starting youtube chat monitor")
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
		logger.Log.Debug("Polling Youtube")

		url := "https://www.googleapis.com/youtube/v3/liveChat/messages?part=snippet,authorDetails&liveChatId=" + liveChatID
		if nextPageToken != "" {
			url += "&pageToken=" + nextPageToken
		}
		_, body, _ := gorequest.New().
			Get(url).
			Set("Authorization", "Bearer "+accessToken).
			End()

		newMessages := []sockets.SocketMessage{}
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

			newMessages = append(newMessages, sockets.SocketMessage{
				Message: message,
				Source:  y.name,
				User:    name,
			})
		}

		for _, socketMessage := range newMessages {
			sockets.SocketChannel <- &socketMessage
		}
		time.Sleep(sleepDuration)
	}
}

// ConnectChat connects to chat service if available
func (y *YoutubeService) ConnectChat() error {
	viper := viper.GetViper()
	if viper.GetString("youtube-livekey") != "" && viper.GetString("youtube-token") != "" {
		go y.connectChat(viper.GetString("youtube-token"))
		return nil
	}
	return errors.New("Missing keys")
}

func init() {
	RegisterService(&YoutubeService{name: "youtube"})
}
