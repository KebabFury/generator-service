package services

import "github.com/KebabFury/generator-service/internal/config"

type Services struct {
	Bots BotService
}

type BotService interface {
}

func NewServices(agnia config.AgniaConfig) *Services {
	return &Services{
		Bots: NewBotServiceImp(agnia),
	}
}
