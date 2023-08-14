package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
)

func main() {
	// Initialize default config
	app := fiber.New()
	app.Use(recover.New(
		recover.Config{
			EnableStackTrace: true,
		}))

	app.Post("/b", HandleB)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("cannot listen on port 8080: %v", err)
	}
}

func HandleB(c *fiber.Ctx) error {
	return c.Send([]byte("hello from B"))
}
