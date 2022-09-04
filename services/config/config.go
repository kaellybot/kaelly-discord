package config

import (
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ConfigService interface {
	GetString(key string) string
	GetInt(key string) int
}

type ConfigServiceImpl struct {
}

func New() (*ConfigServiceImpl, error) {
	viper.SetConfigFile(models.ConfigFileName)

	for configName, defaultValue := range models.DefaultConfigValues {
		viper.SetDefault(configName, defaultValue)
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Str(models.LogFileName, models.ConfigFileName).Msgf("Failed to read config")
		return nil, err
	}

	log.Info().Str(models.LogFileName, models.ConfigFileName).Msgf("Config read!")
	return &ConfigServiceImpl{}, err
}

func (service *ConfigServiceImpl) GetString(key string) string {
	return viper.GetString(key)
}

func (service *ConfigServiceImpl) GetInt(key string) int {
	return viper.GetInt(key)
}
