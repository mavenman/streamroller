//go:generate go-bindata -nomemcopy -pkg main -o ./bindata.go static/...

package main

import (
	"bytes"
	"fmt"
	"mime"
	"net"
	"os"
	"path"
	"strconv"

	"github.com/dustinblackman/streamroller/logger"
	"github.com/dustinblackman/streamroller/services"
	"github.com/dustinblackman/streamroller/sockets"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nareix/bits/pio"
)

var tcpChannel = make(chan []byte, 1)

func writeToConn(conn *net.TCPConn, ctx echo.Context) {
	fmt.Println("URL HIT: ", ctx.Request().URL.String())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(ctx.Request().Body)
	if err != nil {
		fmt.Println("READ FROM BODY ERROR: ", err)
		return
	}

	// fmt.Println("REQUEST: ", buf.Bytes())
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		fmt.Println("Write", err)
	}
}

func channelLoop(conn *net.TCPConn) {
	for {
		fmt.Println("LOOPED")
		// data := make([]byte, 3073)
		data := make([]byte, pio.RecommendBufioSize)
		// data := make([]byte, 8192)
		_, err := conn.Read(data)
		if err != nil {
			fmt.Println("Read", err)
		}

		tcpChannel <- data
	}
}

var emptyMessages = 0
var currentDelay = 1

func readFromConn() []byte {
	fmt.Println("READING FROM CHANNEL")
	if len(tcpChannel) == 0 {
		fmt.Println("CHANNEL IS EMPTY")

		emptyMessages++
		if emptyMessages%10 == 0 {
			fmt.Println("HIT THIS TRESHOLD")
			emptyMessages = 0
			if currentDelay+4 > 21 {
				currentDelay = 21
			} else {
				currentDelay = currentDelay + 4
			}
		}

		// return []byte{0x01}
		return []byte{}
	}

	currentDelay = 1
	emptyMessages = 0

	return <-tcpChannel
}

// func readFromConn(conn *net.TCPConn, ctx echo.Context) error {
// // https://github.com/dustinblackman/joy4/blob/master/format/rtmp/rtmp.go#L1085
// data := make([]byte, 3073) // Play with this number, this I think is the last piece
// _, err := conn.Read(data)
// if err != nil {
// fmt.Println("Read", err)
// }

// fmt.Println("DATA: ", data)
// // fmt.Println("HEADER TYPE: ", data[0]>>6)
// // fmt.Println("OTHER HEADER TYPE: ", data[1]>>6)
// // fmt.Println("CSID: ", uint32(data[0])&0x3f)
// // fmt.Println("HEADER: ", data[:11])
// // fmt.Println("LENGTH BYTES 4-7?: ", data[4:7])
// // fmt.Println("LENGTH BYTES 3-6?: ", data[3:6])
// // fmt.Println("SINGLE LENGTH BYTE?: ", data[4])
// // fmt.Println("PIO 4-7: ", pio.U24BE(data[4:7]))
// // fmt.Println("PIO 3-6", pio.U24BE(data[3:6]))

// // parsed := []byte{}
// // zeroCount := 0
// // for _, b := range data {
// // parsed = append(parsed, b)
// // if b == 0 {
// // zeroCount++
// // } else {
// // zeroCount = 0
// // }

// // if zeroCount == 10 {
// // break
// // }
// // }

// // parsed = parsed[:len(parsed)-10]

// // fmt.Println("PARSED: ", parsed)
// // dst := make([]byte, hex.EncodedLen(len(parsed)))
// // hex.Encode(dst, parsed)

// // fmt.Printf("HEX: %s\n", dst)

// return ctx.Blob(200, "application/x-fcs", data)
// }

// CreateEcho creates the HTTP server.
func CreateEcho(port, rtmpPort string) {
	app := echo.New()
	app.Use(middleware.Gzip())
	app.Use(middleware.CORS())
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// go-bindata middleware
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			uri := ctx.Request().URL.RequestURI()
			data, err := Asset("static" + uri)
			if err != nil {
				return next(ctx)
			}
			return ctx.Blob(200, mime.TypeByExtension(path.Ext(uri)), data)
		}
	})

	app.GET("/", func(ctx echo.Context) error {
		data, err := Asset("static/index.html")
		if err != nil {
			return ctx.String(404, "Not found")
		}
		return ctx.HTML(200, string(data))
	})
	app.GET("/socket", sockets.HandleWebSocketConnections())

	// https://github.com/kwojtek/nginx-rtmpt-proxy-module/blob/master/ngx_rtmpt_proxy_module.c#L486-L504
	var testConn *net.TCPConn
	app.POST("/open/:n", func(ctx echo.Context) error {
		testConn = createLocalConnection(rtmpPort)
		go channelLoop(testConn)
		return ctx.String(200, "12345")
	})
	app.POST("/idle/:session_id/:seq", func(ctx echo.Context) error {
		seq, _ := strconv.Atoi(ctx.Param("seq"))
		if seq == 100 {
			os.Exit(0)
		}

		// if seq < 20 {
		// return ctx.Blob(200, "application/x-fcs", []byte{0x01})
		// }
		// toSend := readFromConn()
		// if len(toSend) == 0 {
		// toSend = []byte{0x01}
		// }

		toSend := []byte{byte(currentDelay)}
		connRes := readFromConn()
		if len(connRes) > 0 {
			toSend = append(toSend, connRes[:len(connRes)-1]...)
		}
		fmt.Println("CURRENT DELAY: ", currentDelay)
		fmt.Println("EMPTY MESSAGES: ", emptyMessages)
		// fmt.Println("TO SEND WITH IDLE: ", hex.Dump(toSend))
		ctx.Response().Header().Add("Connection", "Keep-Alive")
		ctx.Response().Header().Add("Cache-Control", "no-cache")
		ctx.Response().Header().Add("Content-Length", strconv.Itoa(len(toSend)))
		return ctx.Blob(200, "application/x-fcs", toSend)
	})
	// Need to handle handshake, that's why nothing is working
	app.POST("/send/:session_id/:seq", func(ctx echo.Context) error {
		writeToConn(testConn, ctx)
		return ctx.Blob(200, "application/x-fcs", []byte{0x01})
	})

	app.POST("/close/:session_id/:seq", func(ctx echo.Context) error {
		return ctx.Blob(200, "application/x-fcs", []byte{0x01})
		// return readFromConn(testConn, ctx)
	})

	// Setup chat
	for _, service := range services.Services {
		err := service.ConnectChat()
		if err != nil {
			logger.Log.Debug(err)
		}
	}

	logger.Log.Info("Starting Web server")
	go app.Start(":" + port)
}
