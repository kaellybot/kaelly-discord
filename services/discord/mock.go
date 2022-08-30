package discord

import "github.com/rs/zerolog/log"

type DiscordServiceMock struct {
	ListenFunc           func() error
	RegisterCommandsFunc func() error
	ShutdownFunc         func() error
}

func (service *DiscordServiceMock) Listen() error {
	if service.ListenFunc != nil {
		return service.ListenFunc()
	}
	log.Warn().Msgf("Listen is not mocked")
	return nil
}

func (service *DiscordServiceMock) RegisterCommands() error {
	if service.RegisterCommandsFunc != nil {
		return service.RegisterCommandsFunc()
	}
	log.Warn().Msgf("RegisterCommands is not mocked")
	return nil
}

func (service *DiscordServiceMock) Shutdown() error {
	if service.ShutdownFunc != nil {
		return service.ShutdownFunc()
	}
	log.Warn().Msgf("Shutdown is not mocked")
	return nil
}
