package services

import (
	"github.com/KebabFury/generator-service/internal/config"
	"github.com/KebabFury/generator-service/internal/domain"
)

type Services struct {
	Agnia AgniaService
}

type AgniaService interface {
	RegisterProviders(providers []domain.Provider)
}

func NewServices(agnia config.AgniaConfig) *Services {
	return &Services{
		Agnia: NewAgniaServiceImp(agnia),
	}
}
