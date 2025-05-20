package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/proximyst/email-sub/pkg/db"
	"golang.org/x/sync/errgroup"
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
		SQSClient:    sqs.NewFromConfig(cfg),
	}

	lambda.Start(srv.Handler)
}

type Srv struct {
	DynamoClient *dynamodb.Client
	SQSClient    *sqs.Client
}

func (s *Srv) Handler(ctx context.Context, event events.EventBridgeEvent) error {
	db := db.New(s.DynamoClient, os.Getenv("DYNAMODB_TABLE_NAME"))
	feeds, err := db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	errgroup, ctx := errgroup.WithContext(ctx)
	for _, feed := range feeds {
		errgroup.Go(func() error {
			slog.Info("publishing feed to sqs", "feed", feed)
			output, err := s.SQSClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: &feed,
				QueueUrl:    aws.String(os.Getenv("SQS_QUEUE_URL")),
			})
			if err != nil {
				return fmt.Errorf("failed to send message to sqs: %w", err)
			}
			slog.Info("published feed to sqs", "feed", feed, "message_id", *output.MessageId)
			return nil
		})
	}
	if err := errgroup.Wait(); err != nil {
		return fmt.Errorf("failed to publish feeds to sqs: %w", err)
	}

	return nil
}
