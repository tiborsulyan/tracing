package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"log"
	"tracing/cmd/tracing"
)

func main() {
	_, stopFunc := tracing.InitTracer("serviceb", "http://jaeger:14268/api/traces")
	defer stopFunc()

	app := fiber.New()

	app.Use(recover.New(
		recover.Config{
			EnableStackTrace: true,
		}))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} ${reqHeader:traceparent}\n",
	}))
	app.Use(tracing.Middleware())

	app.Post("/b", HandleB)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("cannot listen on port 8080: %v", err)
	}
}

func HandleB(c *fiber.Ctx) error {
	_, span := otel.GetTracerProvider().Tracer("").Start(
		c.UserContext(),
		"HandleB",
	)
	defer span.End()
	return c.Send([]byte("hello from B"))
}
