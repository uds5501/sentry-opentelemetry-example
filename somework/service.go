package somework

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"time"
)

var tracer = otel.Tracer("sub-function")

func MiddleWork(ctx context.Context, id string) {
	ctx2, span := tracer.Start(ctx, "MiddleWork()", oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()
	span.AddEvent("We have started the sleep event")
	time.Sleep(100 * time.Microsecond)
	span.AddEvent("We have completed the sleep event")
	ErrorWork(ctx2, id)
	span.AddEvent("Out of Error!")
}

func ErrorWork(ctx context.Context, id string) {
	_, span := tracer.Start(ctx, "MiddleWork()", oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()
	db, err := sql.Open("mysql", "username:password@tcp(127.0.0.1:3306)/test")
	if err != nil {
		span.RecordError(err)
		return
	}
	span.AddEvent("DB Connection done", oteltrace.WithAttributes(attribute.Any("db", db)))
}
