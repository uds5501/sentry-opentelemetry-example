package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"main/somework"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("gin-server")

func main() {
	log.Println("Welcome to otlp instrumentation service")
	shutdown := initTracer()
	defer shutdown()
	r := gin.New()
	r.Use(otelgin.Middleware("my-server"))

	tmplName := "user"
	tmplStr := "user {{ .name }} (id {{ .id }})\n"
	tmpl := template.Must(template.New(tmplName).Parse(tmplStr))
	r.SetHTMLTemplate(tmpl)

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		name := getUser(c, id)
		oneMore(c)
		otelgin.HTML(c, http.StatusOK, tmplName, gin.H{
			"name": name,
			"id":   id,
		})
	})
	_ = r.Run(":8088")
}

func initTracer() func() {
	ctx := context.Background()
	log.Println("Initialising tracer")
	// cmpExp := component.ExporterCreateSettings{Logger: logger}
	// exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	otelAgentAddr := "otel-collector.observability.svc.cluster.local:4317"
	log.Println("Connecting to GRPC endpoint...")
	exp, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(otelAgentAddr),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	))
	log.Println("Connection established.")
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("test-service"),
		),
	)
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// exporter2, _ := sentry.CreateSentryExporter(&sentry.Config{DSN: "http://2d2364fe1c374dd9bec6ace7e48f82b8@sentry.trell.co/4"}, cmpExp)
	// log.Println(exporter2)
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func() {
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown provider")
		handleErr(exp.Shutdown(ctx), "failed to stop exporter")
	}
}

func getUser(c *gin.Context, id string) string {
	// Pass the built-in `context.Context` object from http.Request to OpenTelemetry APIs
	// where required. It is available from gin.Context.Request.Context()
	ctx, span := tracer.Start(c.Request.Context(), "getUser", oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()
	if id == "123" {
		somework.MiddleWork(ctx, id)
		return "otelgin tester"
	}
	return "unknown"
}
func oneMore(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "apiHit", oteltrace.WithAttributes(attribute.String("test", "onetwothreee")))
	defer span.End()
	app, err := http.Get("https://trell11.co")
	if err != nil {
		span.RecordError(err)
	}
	fmt.Println(app)
}
func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
