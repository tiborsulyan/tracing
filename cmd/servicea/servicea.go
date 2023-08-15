package main

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"strconv"
	"tracing/cmd/tracing"
)

func main() {

	_, stopFunc := tracing.InitTracer("servicea", "http://jaeger:14268/api/traces")
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

	app.Post("/a", HandleA)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("cannot listen on port 8080: %v", err)
	}
}

func HandleA(c *fiber.Ctx) error {
	_, span := otel.GetTracerProvider().Tracer("").Start(
		c.UserContext(),
		"HandleA",
	)
	defer span.End()

	agent := fiber.AcquireAgent()
	defer func() {
		fiber.ReleaseAgent(agent)
	}()
	req := agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	// Inject telemetry context into the outgoing request
	// TODO find a better way to inject into Fiber request headers
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(c.UserContext(), carrier)
	for key, val := range carrier {
		fmt.Printf("key: %v, val: %v\n", key, val)
		req.Header.Set(key, val)
	}
	req.SetRequestURI("http://serviceb:8080/b")
	if err := agent.Parse(); err != nil {
		return err
	}
	var response fiber.Response
	if err := agent.Do(req, &response); err != nil {
		return err
	}

	if response.StatusCode() != fiber.StatusOK {
		return errors.New("serviceb returned" + strconv.Itoa(response.StatusCode()))
	}

	return c.Send([]byte("hello from A\n"))
}
