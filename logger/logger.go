package logger

import "go.uber.org/zap"

var Log *zap.SugaredLogger

func New(debug bool) {
	var initLogger *zap.Logger
	if debug {
		initLogger, _ = zap.NewDevelopment()
	} else {
		initLogger, _ = zap.NewProduction()
	}
	Log = initLogger.Sugar()
}
