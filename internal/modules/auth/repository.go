package auth

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Register(
	ctx context.Context,
	user User,
	household Household,
	member HouseholdMember,
) error {
	session, err := r.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("start MongoDB session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(txCtx context.Context) (any, error) {
			if _, err := r.db.Collection("users").InsertOne(txCtx, user); err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return nil, ErrEmailAlreadyExists
				}
				return nil, fmt.Errorf("insert user: %w", err)
			}

			if _, err := r.db.Collection("households").InsertOne(txCtx, household); err != nil {
				return nil, fmt.Errorf("insert household: %w", err)
			}

			if _, err := r.db.Collection("household_members").InsertOne(txCtx, member); err != nil {
				return nil, fmt.Errorf("insert household member: %w", err)
			}

			return nil, nil
		},
	)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return err
		}
		return fmt.Errorf("register transaction: %w", err)
	}

	return nil
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	var user User

	err := r.db.Collection("users").
		FindOne(ctx, map[string]any{"email": email}).
		Decode(&user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (User, error) {
	var user User

	err := r.db.Collection("users").
		FindOne(ctx, bson.M{"_id": id}).
		Decode(&user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user by ID: %w", err)
	}

	return user, nil
}
