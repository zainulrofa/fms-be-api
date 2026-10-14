package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// Unique index untuk email pengguna.
	users := db.Collection("users")

	_, err := users.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uq_users_email"),
		},
	)
	if err != nil {
		return fmt.Errorf("create users email index: %w", err)
	}

	// Unique index untuk kombinasi householdId dan userId.
	members := db.Collection("household_members")

	_, err = members.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "householdId", Value: 1},
				{Key: "userId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uq_household_members_household_user"),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create household member unique index: %w",
			err,
		)
	}

	// Index untuk pencarian anggota berdasarkan userId.
	_, err = members.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
			Options: options.Index().
				SetName("idx_household_members_user"),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create household member user index: %w",
			err,
		)
	}

	return nil
}
