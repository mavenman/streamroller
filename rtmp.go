package main

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustinblackman/joy4/format/rtmp"
	"github.com/nareix/joy4/av"
	"github.com/spf13/viper"
)

func addRTMPConnection(rtmps []*rtmp.Conn, url string) []*rtmp.Conn {
	conn, err := rtmp.Dial(url)
	if err != nil {
		logrus.Panic(err)
	}

	return append(rtmps, conn)
}

func copyPackets(src av.PacketReader, rtmps []*rtmp.Conn) (err error) {
	var pkgChans []chan av.Packet
	for _, conn := range rtmps {
		pktChan := make(chan av.Packet)
		pkgChans = append(pkgChans, pktChan)

		go func(conn *rtmp.Conn, pkgChan <-chan av.Packet) {
			for pkt := range pkgChan {
				if err = conn.WritePacket(pkt); err != nil {
					return
				}
			}
		}(conn, pktChan)
	}

	sourceChan := make(chan av.Packet, 1)
	errorChan := make(chan error, 1)
	go func() {
		for {
			var pkt av.Packet
			if pkt, err = src.ReadPacket(); err != nil {
				errorChan <- err
				break
			} else {
				sourceChan <- pkt
			}
		}
	}()

	for {
		select {
		case pkt := <-sourceChan:
			for _, pkgChan := range pkgChans {
				pkgChan <- pkt
			}
		case err = <-errorChan:
			return
		case <-time.After(time.Second * 20):
			err = errors.New("Packet timeout reached")
			return
		}
	}
}

func writeHeaders(src av.Demuxer, rtmps []*rtmp.Conn) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}

	for _, conn := range rtmps {
		if err = conn.WriteHeader(streams); err != nil {
			return
		}
	}

	return
}

func closeConnections(rtmps []*rtmp.Conn) (err error) {
	for _, conn := range rtmps {
		if err = conn.WriteTrailer(); err != nil {
			return
		}
		conn.Close()
	}
	return
}

func handlePublish(conn *rtmp.Conn) {
	// fmt.Println(conn.URL) // TODO: Add stream key verification
	viper := viper.GetViper()
	log := logrus.WithFields(logrus.Fields{"module": "rtmp"})

	// Handles creating RTMP connections to services
	var rtmps []*rtmp.Conn
	if viper.GetString("twitch-livekey") != "" {
		log.Debug("Dialing Twitch")
		rtmps = addRTMPConnection(rtmps, "rtmp://live.twitch.tv/app/"+viper.GetString("twitch-livekey"))
	}
	if viper.GetString("facebook-livekey") != "" {
		log.Debug("Dialing Facebook")
		rtmps = addRTMPConnection(rtmps, "rtmp://rtmp-api.facebook.com:80/rtmp/"+viper.GetString("facebook-livekey"))
	}
	if viper.GetString("youtube-livekey") != "" {
		log.Debug("Dialing Youtube")
		rtmps = addRTMPConnection(rtmps, "rtmp://a.rtmp.youtube.com/live2/"+viper.GetString("youtube-livekey"))
	}

	err := writeHeaders(conn, rtmps)
	if err != nil {
		log.WithFields(logrus.Fields{"func": "writeHeaders"}).Error(err)
	}
	err = copyPackets(conn, rtmps)
	if err != nil {
		log.WithFields(logrus.Fields{"func": "copyPackets"}).Error(err)
	}
	err = closeConnections(rtmps)
	if err != nil {
		log.WithFields(logrus.Fields{"func": "closeConnections"}).Error(err)
	}
	conn.Close()
}

// CreateRTMP creates a new RTMP server
func CreateRTMP(port string) {
	server := &rtmp.Server{Addr: ":" + port}
	server.HandlePublish = handlePublish

	logrus.WithFields(logrus.Fields{"module": "rtmp"}).Info("Starting RTMP")
	go server.ListenAndServe()
}
