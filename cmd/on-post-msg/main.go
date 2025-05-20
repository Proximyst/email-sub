package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/proximyst/email-sub/pkg/db"
	"github.com/proximyst/email-sub/pkg/models"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
		panic("unreachable")
	}

	srv := &Srv{
		DynamoClient: dynamodb.NewFromConfig(cfg),
		SESClient:    ses.NewFromConfig(cfg),
	}

	lambda.Start(srv.Handler)
}

type Srv struct {
	DynamoClient *dynamodb.Client
	SESClient    *ses.Client
}

func (s *Srv) Handler(ctx context.Context, event events.SQSEvent) error {
	if len(event.Records) == 0 {
		return nil // wtf?
	} else if len(event.Records) > 1 {
		return fmt.Errorf("expected 1 record, got %d", len(event.Records))
	}

	data := db.New(s.DynamoClient, os.Getenv("DYNAMODB_TABLE_NAME"))
	msg := event.Records[0]
	var request models.FeedPostEmailRequest
	if err := json.Unmarshal([]byte(msg.Body), &request); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	err := data.InsertEmailSent(ctx, request.Feed, request.ID, request.Email)
	if errors.Is(err, db.ErrCellAlreadyExists) {
		slog.Info("email already sent", "feed", request.Feed, "email", request.Email)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert email sent: %w", err)
	}

	// TODO: Send with SES

	slog.Info("email sent", "feed", request.Feed, "email", request.Email)

	return nil
}
