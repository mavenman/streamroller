package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/dustinblackman/streamroller/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "HEAD"
)

// Server holds variables for working with the HTTP and RTMP server
type Server struct {
	HTTPPort string
	RTMPPort string
}

// getPort checks for a random available port on system and verifies we can listen on it before returning.
func getPort() string {
	var port int
	for {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			panic(err)
		}

		listen, err := net.ListenTCP("tcp", addr)
		if err != nil {
			logger.Log.Error(err)
			continue
		}

		port = listen.Addr().(*net.TCPAddr).Port
		listen.Close()
		break
	}

	return strconv.Itoa(port)
}

func createLocalConnection(port string) *net.TCPConn {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:"+port)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}
	return conn
}

// ProxyConnection verifies whether connection is RTMP or HTTP, and redirects traffic accordingly.
func (s *Server) ProxyConnection(conn *net.TCPConn) {
	defer conn.Close()
	data := make([]byte, 1)
	n, err := conn.Read(data)
	if err != nil {
		logger.Log.Error(err)
		return
	}

	var proxyConn *net.TCPConn
	if data[0] == 0x03 { // RTMP first byte.
		logger.Log.Debug("Forwarding RTMP connection")
		proxyConn = createLocalConnection(s.RTMPPort)
	} else {
		logger.Log.Debug("Forwarding HTTP connection")
		proxyConn = createLocalConnection(s.HTTPPort)
	}
	proxyConn.Write(data[:n])
	defer proxyConn.Close()

	// Request loop
	go func() {
		for {
			data := make([]byte, 1024*1024)
			n, err := conn.Read(data)
			if err != nil {
				logger.Log.Debug(err)
				break
			}
			proxyConn.Write(data[:n])
		}
	}()

	// Response loop
	for {
		data := make([]byte, 1024*1024)
		n, err := proxyConn.Read(data)
		if err != nil {
			logger.Log.Debug(err)
			break
		}
		conn.Write(data[:n])
	}
}

func run(rootCmd *cobra.Command, args []string) {
	viper.AutomaticEnv()
	viper.ReadInConfig()

	// Setup global logger
	logger.New(viper.GetBool("debug"))

	server := Server{getPort(), "9090"}
	CreateEcho(server.HTTPPort, server.RTMPPort)
	CreateRTMP(server.RTMPPort)

	addr, err := net.ResolveTCPAddr("tcp", ":"+viper.GetString("port"))
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	logger.Log.Infof("Starting to listen for connections on port %s", viper.GetString("port"))
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go server.ProxyConnection(conn)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "streamroller",
		Example: `  streamroller -t TWITCH-KEY -f FACEBOOK-KEY -y YOUTUBE-KEY`,
		Run:     run,
		Short:   "A multi streaming tool for with read only merged chats for platforms like Twitch, Youtube, and Facebook",
		Long: `A multi streaming tool for with read only merged chats for platforms like Twitch, Youtube, and Facebook

Version: ` + version + `
Home: https://github.com/dustinblackman/streamroller`,
	}

	flags := rootCmd.PersistentFlags()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	flags.String("port", port, "Server port")
	flags.Bool("debug", false, "Enable debug logging")

	// Facebook
	flags.StringP("facebook-livekey", "f", "", "Facebook live stream key")
	flags.String("facebook-token", "", "Facebook access token")

	// Twitch
	flags.StringP("twitch-livekey", "t", "", "Twitch live key")
	flags.String("twitch-username", "", "Twitch channel user name")
	flags.String("twitch-oauth", "", "Twitch oauth key. It can be generated here: https://twitchapps.com/tmi/")

	// Youtube
	flags.StringP("youtube-livekey", "y", "", "Youtube live key")
	flags.String("youtube-token", "", "Youtube refresh token")

	viper.SetConfigName("streamroller")
	viper.SetEnvPrefix("sr")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("$HOME/.streamroller")

	for _, param := range []string{
		"port",
		"debug",
		"facebook-livekey",
		"facebook-token",
		"twitch-livekey",
		"twitch-username",
		"twitch-oauth",
		"youtube-livekey",
		"youtube-token"} {
		viper.BindPFlag(param, flags.Lookup(param))
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
