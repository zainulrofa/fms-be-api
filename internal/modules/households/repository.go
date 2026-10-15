package households

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

// FindByUserID mengambil semua household yang diikuti pengguna.
func (r *Repository) FindByUserID(
	ctx context.Context,
	userID bson.ObjectID,
) ([]Household, error) {
	cursor, err := r.db.Collection("household_members").Find(
		ctx,
		bson.M{"userId": userID},
	)
	if err != nil {
		return nil, fmt.Errorf("find household memberships: %w", err)
	}
	defer cursor.Close(ctx)

	var memberships []struct {
		HouseholdID bson.ObjectID `bson:"householdId"`
	}

	if err := cursor.All(ctx, &memberships); err != nil {
		return nil, fmt.Errorf("decode household memberships: %w", err)
	}

	households := make([]Household, 0, len(memberships))
	if len(memberships) == 0 {
		return households, nil
	}

	householdIDs := make([]bson.ObjectID, 0, len(memberships))
	for _, membership := range memberships {
		householdIDs = append(householdIDs, membership.HouseholdID)
	}

	householdCursor, err := r.db.Collection(CollectionName).Find(
		ctx,
		bson.M{"_id": bson.M{"$in": householdIDs}},
	)
	if err != nil {
		return nil, fmt.Errorf("find households: %w", err)
	}
	defer householdCursor.Close(ctx)

	if err := householdCursor.All(ctx, &households); err != nil {
		return nil, fmt.Errorf("decode households: %w", err)
	}

	return households, nil
}

// IsMember memeriksa apakah pengguna merupakan anggota household.
func (r *Repository) IsMember(
	ctx context.Context,
	householdID bson.ObjectID,
	userID bson.ObjectID,
) (bool, error) {
	err := r.db.Collection("household_members").FindOne(
		ctx,
		bson.M{
			"householdId": householdID,
			"userId":      userID,
		},
	).Err()

	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check household membership: %w", err)
	}

	return true, nil
}

// FindByID mengambil household berdasarkan ID.
func (r *Repository) FindByID(
	ctx context.Context,
	householdID bson.ObjectID,
) (Household, error) {
	var household Household

	err := r.db.Collection(CollectionName).FindOne(
		ctx,
		bson.M{"_id": householdID},
	).Decode(&household)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return Household{}, mongo.ErrNoDocuments
	}
	if err != nil {
		return Household{}, fmt.Errorf("find household by ID: %w", err)
	}

	return household, nil
}

// FindMembers mengambil daftar anggota suatu household.

func (r *Repository) FindMembers(
	ctx context.Context,
	householdID bson.ObjectID,
) ([]MemberResponse, error) {
	members := make([]MemberResponse, 0)

	cursor, err := r.db.Collection("household_members").Find(
		ctx,
		bson.M{"householdId": householdID},
	)
	if err != nil {
		return nil, fmt.Errorf("find household members: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &members); err != nil {
		return nil, fmt.Errorf("decode household members: %w", err)
	}

	return members, nil
}

