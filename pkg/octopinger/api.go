package octopinger

import (
	"context"

	srv "github.com/katallaxie/pkg/server"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type api struct {
	srv.Listener
}

// NewAPI ...
func NewAPI() *api {
	a := new(api)

	return a
}

// Start ...
func (a *api) Start(ctx context.Context, ready srv.ReadyFunc, run srv.RunFunc) func() error {
	return func() error {
		app := fiber.New()

		app.Use(recover.New())
		app.Use(requestid.New())
		app.Use(logger.New())

		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hello, World 🐙!")
		})

		app.Get("/health", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		go func() {
			<-ctx.Done()
			_ = app.Shutdown()
		}()

		err := app.Listen(":3000")
		if err != nil {
			return err
		}

		return err
	}
}
