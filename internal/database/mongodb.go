package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	if uri == "" {
		return nil, errors.New("MONGODB_URI is not configured")
	}

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		return nil, fmt.Errorf("create MongoDB client: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()

		if disconnectErr := client.Disconnect(cleanupCtx); disconnectErr != nil {
			return nil, fmt.Errorf(
				"ping MongoDB: %w; cleanup failed: %v",
				err,
				disconnectErr,
			)
		}

		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}
