package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/mmcdole/gofeed"
	"github.com/proximyst/email-sub/pkg/batching"
	"github.com/proximyst/email-sub/pkg/db"
	"github.com/proximyst/email-sub/pkg/ids"
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
		SQSClient:    sqs.NewFromConfig(cfg),
	}

	lambda.Start(srv.Handler)
}

type Srv struct {
	DynamoClient *dynamodb.Client
	SQSClient    *sqs.Client
}

func (s *Srv) Handler(ctx context.Context, event events.SQSEvent) error {
	if len(event.Records) == 0 {
		return nil // wtf?
	} else if len(event.Records) > 1 {
		return fmt.Errorf("expected 1 record, got %d", len(event.Records))
	}

	data := db.New(s.DynamoClient, os.Getenv("DYNAMODB_TABLE_NAME"))
	msg := event.Records[0]
	feedURL := msg.Body

	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(feedURL)
	if err != nil {
		return fmt.Errorf("failed to parse feed: %w", err)
	}

	cells := make([]db.FeedPostCell, 0, len(feed.Items))
	for _, item := range feed.Items {
		id := item.GUID
		if id == "" {
			id = item.Link
		}
		if id == "" {
			continue
		}

		posted := time.Now()
		if item.PublishedParsed != nil {
			posted = *item.PublishedParsed
		}

		link := item.Link
		if link == "" && len(item.Links) > 0 {
			link = item.Links[0]
		}
		if link == "" && (strings.HasPrefix(item.GUID, "https://") || strings.HasPrefix(item.GUID, "http://")) {
			link = item.GUID
		}
		if link == "" {
			continue
		}

		cells = append(cells, db.FeedPostCell{
			Feed:   feedURL,
			ID:     id,
			Posted: posted,
			Link:   link,
		})
	}

	cells, err = data.FilterExistingPosts(ctx, feedURL, cells)
	if err != nil {
		return fmt.Errorf("failed to filter existing posts: %w", err)
	}
	if len(cells) == 0 {
		slog.Info("no new posts found", "feed", feedURL)
		return nil
	}

	subscriptions, err := data.GetSubscriptionsForFeed(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for feed: %w", err)
	}

	sqsURL := os.Getenv("SQS_QUEUE_URL")
	for _, cell := range cells {
		prototype := models.FeedPostEmailRequest{
			Feed:   cell.Feed,
			ID:     cell.ID,
			Link:   cell.Link,
			Posted: cell.Posted,
		}
		requests := make([]types.SendMessageBatchRequestEntry, 0, len(subscriptions))
		for _, sub := range subscriptions {
			req := prototype
			req.Email = sub.Email
			reqJSON, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			requests = append(requests, types.SendMessageBatchRequestEntry{
				Id:          aws.String(ids.CalculateID(cell.Feed + "#" + cell.ID + "#" + sub.Email)),
				MessageBody: aws.String(string(reqJSON)),
			})
		}

		for _, batch := range batching.Batch(requests, 10) {
			if _, err := s.SQSClient.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
				QueueUrl: &sqsURL,
				Entries:  batch,
			}); err != nil {
				return fmt.Errorf("failed to send message batch: %w", err)
			}
		}

		if err := data.InsertFeedPost(ctx, cell); err != nil {
			return fmt.Errorf("failed to insert feed post: %w", err)
		}
		slog.Info("processed feed post", "feed", cell.Feed, "id", cell.ID, "link", cell.Link)
	}

	return nil
}
