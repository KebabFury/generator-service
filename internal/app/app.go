package app

import (
	"context"
	_ "github.com/KebabFury/generator-service/docs"
	"github.com/KebabFury/generator-service/internal/config"
	"github.com/KebabFury/generator-service/internal/handlers/http"
	"github.com/KebabFury/generator-service/internal/services"
	"github.com/KebabFury/generator-service/pkg/logger"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"os/signal"
	"syscall"
)

// @title API
// @version 1.0
// @description API
// @host localhost:8000
// @BasePath /api
// Run initializes whole application.
func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	l := logger.Get()
	conf, err := config.Init()
	if err != nil {
		l.Info().Err(err).Msg("cannot load config")
	}

	engine := html.New("./py", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	s := services.NewServices(conf.Agnia)
	logger := logger.NewConsole()
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	h := http.NewHandler(s)
	h.Init(app)

	go func() {
		if err := app.Listen(":8000"); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
	l.Info().Msg("--- Shutdown service ---")
	err = app.Shutdown()
	if err != nil {
		l.Error().Err(err).Msg("Can not stop server")
	}
	l.Info().Msg("--- Service is shutdown ---")
}
