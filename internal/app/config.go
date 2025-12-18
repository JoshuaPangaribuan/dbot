package app

import (
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/config"
)

func initConfig(filenames ...string) (config.Config, error) {
	cfg, err := config.New(
		config.WithFiles(filenames...),
		config.WithWatcher(true),
		config.WithProvider(config.ProviderViper),
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
