package http

import (
	"bytes"
	"github.com/KebabFury/generator-service/pkg/parser"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"regexp"
)

var preRe = regexp.MustCompile(`(?miU)<pre>((?:.|\n)+)<\/pre>`)

func (h *Handler) RegisterProvider(c *fiber.Ctx) error {

	resp, err := http.DefaultClient.Post("http://localhost:8000/api/generate", "text/plain", bytes.NewReader(c.BodyRaw()))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	match := preRe.FindStringSubmatch(string(body))

	pythonFile := match[1]

	h.services.Agnia.RegisterProvider(c.Query("provider"), string(c.Body()), pythonFile)

	return c.Render("actions", parser.ParseDocument(string(c.Body())))
}
