package http

import (
	"encoding/json"
	"github.com/KebabFury/generator-service/internal/domain"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"regexp"
)

var preRe = regexp.MustCompile(`(?miU)<pre>((?:.|\n)+)<\/pre>`)

func (h *Handler) RegisterProvider(c *fiber.Ctx) error {

	var providers []domain.Provider
	resp, err := http.DefaultClient.Get("http://91.197.98.50:5243/provider/list-docs")
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &providers)
	if err != nil {
		return err
	}

	h.services.Agnia.RegisterProviders(providers)

	return c.SendStatus(2000)
}
