//go:generate go-bindata -nomemcopy -pkg streamrollerlib -o ./bindata.go ../static/...

package streamrollerlib

import (
	"mime"
	"path"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/laice/streamroller/logger"
	"github.com/laice/streamroller/services"
	"github.com/laice/streamroller/sockets"
)

// CreateEcho creates the HTTP server.
func CreateEcho(port string) {
	app := echo.New()
	app.Use(middleware.Gzip())
	app.Use(middleware.CORS())
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	// go-bindata middleware
	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			uri := ctx.Request().URL.RequestURI()
			data, err := Asset("../static" + uri)
			if err != nil {
				return next(ctx)
			}
			return ctx.Blob(200, mime.TypeByExtension(path.Ext(uri)), data)
		}
	})

	app.GET("/", func(ctx echo.Context) error {
		data, err := Asset("../static/index.html")
		if err != nil {
			return ctx.String(404, "Not found")
		}
		return ctx.HTML(200, string(data))
	})
	app.GET("/socket", sockets.HandleWebSocketConnections())

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
