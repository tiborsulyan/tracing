package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
	"tracing/cmd/tracing"
)

var tracer trace.Tracer

func main() {

	tp, err := tracing.TracerProvider("servicea", "http://jaeger:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// register propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer = otel.Tracer("servicea")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	// Initialize default config
	app := fiber.New()
	app.Use(recover.New(
		recover.Config{
			EnableStackTrace: true,
		}))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path} ${reqHeader:traceparent}\n",
	}))
	app.Post("/a", HandleA)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("cannot listen on port 8080: %v", err)
	}
}

func HandleA(c *fiber.Ctx) error {
	//agent := fiber.AcquireAgent()
	//defer func() {
	//	fiber.ReleaseAgent(agent)
	//}()
	//req := agent.Request()
	//req.Header.SetMethod(fiber.MethodPost)
	//req.SetRequestURI("http://serviceb/b")
	//var response fiber.Response
	//if err := agent.Do(req, &response); err != nil {
	//	return err
	//}
	//
	//if response.StatusCode() != fiber.StatusOK {
	//	return errors.New("serviceb returned" + strconv.Itoa(response.StatusCode()))
	//}

	// Extract parent span from request headers
	ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.MapCarrier(c.GetReqHeaders()))
	_, span := otel.GetTracerProvider().Tracer("servicea").Start(
		context.Background(),
		"handle",
		trace.WithLinks(trace.LinkFromContext(ctx)),
	)
	defer span.End()
	return c.Send([]byte("hello from A\n"))
}
