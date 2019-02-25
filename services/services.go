package services

import (
	"github.com/laice/joy4/format/rtmp"
	"github.com/laice/streamroller/logger"
)

// Service is the interface for Services
type Service interface {
	CanConnect() bool
	ConnectChat() error
	ConnectRTMP() *rtmp.Conn
	Name() string
}

// Services is an accessible export to list all supported services
var Services []Service
var ServiceNames map[string]Service

func connectRTMP(url string) *rtmp.Conn {
	conn, err := rtmp.Dial(url)
	if err != nil {
		logger.Log.Error(err)
	}

	return conn
}

// RegisterService is called on init for all services
func RegisterService(service Service) {
	Services = append(Services, service)

	ServiceNames[service.Name()] = service
}
