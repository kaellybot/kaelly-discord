package config

import "github.com/rs/zerolog/log"

type ConfigServiceMock struct {
	GetStringFunc func(key string) string
	GetIntFunc    func(key string) int
}

func (service *ConfigServiceMock) GetString(key string) string {
	if service.GetStringFunc != nil {
		return service.GetStringFunc(key)
	}
	log.Warn().Msgf("GetString is not mocked")
	return ""
}

func (service *ConfigServiceMock) GetInt(key string) int {
	if service.GetIntFunc != nil {
		return service.GetIntFunc(key)
	}
	log.Warn().Msgf("GetInt is not mocked")
	return 0
}
