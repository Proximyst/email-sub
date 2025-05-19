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

func Handler(ctx context.Context, req events.LambdaFunctionURLRequest) error {
	slog.InfoContext(ctx, "Lambda Function URL event", "event", req)

	return nil
}
