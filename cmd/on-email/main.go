package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, event events.SimpleEmailEvent) error {
	for _, record := range event.Records {
		ses := record.SES
		slog.InfoContext(ctx, "SES event", "event", ses)
	}
	return nil
}
