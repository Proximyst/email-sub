package db

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/proximyst/email-sub/pkg/batching"
)

// RecordType is used as a prefix in the partition key to identify the type of record inside.
type RecordType string

const (
	RecordTypeFeedSubscription RecordType = "efs"
	RecordTypeFeedPost         RecordType = "fpo"
	RecordTypeEmailSent        RecordType = "ems"
)

func (r RecordType) MarshalText() ([]byte, error) {
	return []byte(r), nil
}

func (r *RecordType) UnmarshalText(data []byte) error {
	*r = RecordType(data)
	switch *r {
	case RecordTypeFeedSubscription:
		return nil
	default:
		return fmt.Errorf("invalid record type: %s", string(data))
	}
}

func (r RecordType) IsKey(key string) bool {
	return strings.HasPrefix(key, string(r)+"#")
}

type FeedSubscriptionCell struct {
	// Feed is used in the partition key. It is a URL to the feed.
	Feed string `json:"feed"`
	// Email is used in the sort key. It is the email address of the user.
	Email string `json:"email"`
}

type FeedPostCell struct {
	// Feed is used in the partition key. It is a URL to the feed.
	Feed string `json:"feed"`
	// ID is used in the sort key. It is the ID of the post, as defined by the feed.
	ID string `json:"id"`

	// Link is the URL to the post. It is used to send the email.
	Link string `json:"link"`
	// Posted is when the post was published. It is used to determine the order of posts.
	// If the timestamp in the feed is not available or invalid, it should be set to the time of finding it.
	Posted time.Time `json:"posted"`
}

var ErrCellAlreadyExists = fmt.Errorf("cell already exists")

type Database interface {
	GetFeeds(ctx context.Context) ([]string, error)
	InsertFeedSubscription(ctx context.Context, cell FeedSubscriptionCell) error
	GetSubscriptionsForFeed(ctx context.Context, feed string) ([]FeedSubscriptionCell, error)
	FilterExistingPosts(ctx context.Context, feed string, cells []FeedPostCell) ([]FeedPostCell, error)
	InsertFeedPost(ctx context.Context, cell FeedPostCell) error
}

var _ Database = (*dynamo)(nil)

type dynamo struct {
	Client *dynamodb.Client
	Table  string
}

func New(client *dynamodb.Client, table string) *dynamo {
	return &dynamo{
		Client: client,
		Table:  table,
	}
}

func (d *dynamo) GetFeeds(ctx context.Context) ([]string, error) {
	output, err := d.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(d.Table),
		FilterExpression: aws.String("begins_with(pk, :pk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: string(RecordTypeFeedSubscription) + "#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan table: %w", err)
	}

	feeds := make(map[string]struct{}, 32)
	for _, item := range output.Items {
		if item["feed"] == nil {
			return nil, fmt.Errorf("feed not found in feed subscription item")
		}
		feed, ok := item["feed"].(*types.AttributeValueMemberS)
		if !ok {
			return nil, fmt.Errorf("feed is not a string")
		}
		feeds[feed.Value] = struct{}{}
	}
	return slices.Collect(maps.Keys(feeds)), nil
}

func (d *dynamo) InsertFeedSubscription(ctx context.Context, cell FeedSubscriptionCell) error {
	_, err := d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.Table),
		Item: map[string]types.AttributeValue{
			"pk":   &types.AttributeValueMemberS{Value: string(RecordTypeFeedSubscription) + "#" + cell.Feed},
			"sk":   &types.AttributeValueMemberS{Value: cell.Email},
			"feed": &types.AttributeValueMemberS{Value: cell.Feed},
		},
	})
	return err
}

func (d *dynamo) GetSubscriptionsForFeed(ctx context.Context, feed string) ([]FeedSubscriptionCell, error) {
	output, err := d.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.Table),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: string(RecordTypeFeedSubscription) + "#" + feed},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}

	cells := make([]FeedSubscriptionCell, 0, len(output.Items))
	for _, item := range output.Items {
		cell := FeedSubscriptionCell{
			Feed:  feed,
			Email: item["sk"].(*types.AttributeValueMemberS).Value,
		}
		cells = append(cells, cell)
	}
	return cells, nil
}

func (d *dynamo) FilterExistingPosts(ctx context.Context, feed string, cells []FeedPostCell) ([]FeedPostCell, error) {
	// Return only the posts that are not already in the database.
	if len(cells) == 0 {
		return nil, nil
	}

	batches := batching.Batch(cells, 30)

	// Query the database for each batch.
	existing := make(map[string]struct{}, len(cells))
	for _, batch := range batches {
		keys := make([]map[string]types.AttributeValue, 0, len(batch))
		for _, cell := range batch {
			keys = append(keys, map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: string(RecordTypeFeedPost) + "#" + feed},
				"sk": &types.AttributeValueMemberS{Value: cell.ID},
			})
		}

		output, err := d.Client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				d.Table: {
					Keys: keys,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to batch get items: %w", err)
		}

		for _, item := range output.Responses[d.Table] {
			existing[item["sk"].(*types.AttributeValueMemberS).Value] = struct{}{}
		}
	}

	// Filter the cells to only include the ones that are not in the database.
	filtered := make([]FeedPostCell, 0, 16)
	for _, cell := range cells {
		if _, ok := existing[cell.ID]; !ok {
			filtered = append(filtered, cell)
		}
	}
	return filtered, nil
}

func (d *dynamo) InsertFeedPost(ctx context.Context, cell FeedPostCell) error {
	_, err := d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.Table),
		Item: map[string]types.AttributeValue{
			"pk":     &types.AttributeValueMemberS{Value: string(RecordTypeFeedPost) + "#" + cell.Feed},
			"sk":     &types.AttributeValueMemberS{Value: cell.ID},
			"link":   &types.AttributeValueMemberS{Value: cell.Link},
			"posted": &types.AttributeValueMemberS{Value: cell.Posted.Format(time.RFC3339)},
		},
	})
	return err
}

func (d *dynamo) InsertEmailSent(ctx context.Context, feed, postID, email string) error {
	_, err := d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.Table),
		Item: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: string(RecordTypeEmailSent) + "#" + feed + "#" + postID},
			"sk": &types.AttributeValueMemberS{Value: email},
		},
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	})
	var conditionalFailed *types.ConditionalCheckFailedException
	if errors.As(err, &conditionalFailed) {
		return ErrCellAlreadyExists
	}
	return err
}
