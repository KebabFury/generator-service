package http

import (
	"bytes"
	"github.com/KebabFury/generator-service/internal/services"
	"github.com/KebabFury/generator-service/pkg/parser"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"io"
	"net/http"
)

type Handler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init(app *fiber.App) {
	app.Use(cors.New())
	app.Get("ui", h.Ui)
	api := app.Group("/api")
	api.Get("/swagger/*", swagger.HandlerDefault) // default

	api.Get("/ping", h.Ping)
	api.Post("/python", h.GeneratePython)
	api.Post("/generate", h.GeneratePythonHtml)
	api.Post("/register", h.RegisterProvider)
	api.Post("/parse/body", h.ParseFromBody)

}

func (h *Handler) Ui(c *fiber.Ctx) error {
	return c.Render("ui", fiber.Map{})
}

// Ping
// @Summary Ping
// @Tags service
// @Description Ping
// @ModuleID Зштп
// @Accept  json
// @Produce  json
//
//	@Success 200 {object} responses.PingResponse
//
// @Failure 400,401,500,503 {null} null
// @Router /ping [get]
func (h *Handler) Ping(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":      true,
		"message": "pong",
	})
}

func (h *Handler) GeneratePython(c *fiber.Ctx) error {
	resp, err := http.DefaultClient.Post("http://localhost:8000/api/generate", "text/plain", bytes.NewReader(c.BodyRaw()))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	match := preRe.FindStringSubmatch(string(body))
	return c.SendString(match[1])
}

func (h *Handler) GeneratePythonHtml(c *fiber.Ctx) error {
	doc := parser.ParseDocument(string(c.Body()))
	doc.Provider = c.Query("provider")
	doc.Description = c.Query("description")
	return c.Render("actions", doc)
}
