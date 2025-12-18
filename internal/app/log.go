package app

import (
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

func initLogger() logger.Logger {
	logger := logger.New(logger.WithProvider(logger.ProviderSlog))
	return logger
}
