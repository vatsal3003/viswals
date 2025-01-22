package logger

import (
	"os"

	"github.com/vatsal3003/viswals/internal/consts"
	"go.uber.org/zap"
)

// New will initialize logger with log level and stack trace config
func New() *zap.Logger {
	logLevel := os.Getenv(consts.LogLevel)

	var logger *zap.Logger
	var err error

	// Initialize logger according to defined log level
	if logLevel == consts.LogLevelDebug {
		logger, err = zap.NewDevelopment(zap.AddStacktrace(zap.DPanicLevel))
	} else {
		logger, err = zap.NewProduction(zap.AddStacktrace(zap.DPanicLevel))
	}

	if err != nil {
		// If logger initialization fails then use the default production logger
		logger, _ = zap.NewProduction(zap.AddStacktrace(zap.DPanicLevel))
	}

	return logger
}
