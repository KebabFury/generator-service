package services

import "github.com/KebabFury/generator-service/internal/config"

type Services struct {
	Agnia AgniaService
}

type AgniaService interface {
	RegisterProvider(provider string, document string, pythonFileContents string)
}

func NewServices(agnia config.AgniaConfig) *Services {
	return &Services{
		Agnia: NewAgniaServiceImp(agnia),
	}
}
