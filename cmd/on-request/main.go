package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	slog.InfoContext(ctx, "Lambda Function URL event", "event", req)

	return events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Body:       req.Body,
	}, nil
}
