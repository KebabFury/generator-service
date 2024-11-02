package services

import "github.com/KebabFury/generator-service/internal/config"

type BotServiceImp struct {
	conf config.AgniaConfig
}

func NewBotServiceImp(agnia config.AgniaConfig) *BotServiceImp {
	return &BotServiceImp{
		agnia,
	}
}
