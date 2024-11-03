package http

import (
	"github.com/KebabFury/generator-service/pkg/parser"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"strings"
)

// Parse
// @Summary Parse swagger to doc and python
// @Tags service
// @Description Parse swagger to doc and python
// @ModuleID 3
// @Accept text/plain
// @Param provider query string true "Provider name"
// @Param data body string true "Input text data"
// @Produce text/plain
// @Success 200 {string} string "Successfully generated python"
// @Failure 400,401,500,503 {string} string "Error occurred"
// @Router /parse [post]
func (h *Handler) Parse(c *fiber.Ctx) error {

	swagParser := parser.NewSwaggerParser()
	doc := swagParser.Parse(c.BodyRaw())
	markdown := parser.DocumentToMarkdown(doc)

	resp, err := http.DefaultClient.Post("http://localhost:8000/api/generate?provider="+c.Query("provider"), "text/plain", strings.NewReader(markdown))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	match := preRe.FindStringSubmatch(string(body))

	pythonFile := match[1]

	return c.JSON(fiber.Map{
		"actions":       pythonFile,
		"documentation": markdown,
	})
}
