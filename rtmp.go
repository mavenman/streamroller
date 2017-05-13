package main

import (
	"errors"
	"time"

	"github.com/dustinblackman/joy4/format/rtmp"
	"github.com/dustinblackman/streamroller/logger"
	"github.com/dustinblackman/streamroller/services"
	"github.com/nareix/joy4/av"
)

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
		case <-time.After(time.Second * 8):
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

	// Handles creating RTMP connections to services
	var rtmps []*rtmp.Conn
	for _, service := range services.Services {
		if service.CanConnect() {
			rtmps = append(rtmps, service.ConnectRTMP())
		}
	}

	err := writeHeaders(conn, rtmps)
	if err != nil {
		logger.Log.Error(err)
	}
	err = copyPackets(conn, rtmps)
	if err != nil {
		logger.Log.Error(err)
	}
	err = closeConnections(rtmps)
	if err != nil {
		logger.Log.Error(err)
	}
	conn.Close()
}

// CreateRTMP creates a new RTMP server
func CreateRTMP(port string) {
	server := &rtmp.Server{Addr: ":" + port}
	server.HandlePublish = handlePublish

	logger.Log.Info("Starting RTMP server")
	go server.ListenAndServe()
}
