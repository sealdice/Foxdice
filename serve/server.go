package serve

import (
	"foxdice/ui"
	"foxdice/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	web *echo.Echo
)

func Serve(config utils.IConfig, logger utils.ILogger) {
	web = echo.New()
	web.HideBanner = true
	web.HidePort = true
	c := middleware.DefaultLoggerConfig
	c.Output = eLog{logger}
	web.Use(middleware.LoggerWithConfig(c))
	web.Use(middleware.Recover())

	web.StaticFS("/", ui.Main)
	web.StaticFS("/bots", ui.Bots)

	port := config.String("server.port")
	logger.Infof("服务开启于 %s", port)
	logger.Fatal(web.Start(":" + port))
}

type eLog struct {
	log utils.ILogger
}

func (e eLog) Write(p []byte) (n int, err error) {
	e.log.Debug(string(p))
	return len(p), nil
}
