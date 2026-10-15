
package households

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrInvalidID       = errors.New("invalid household ID")
	ErrHouseholdNotFound = errors.New("household not found")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListByUser(
	ctx context.Context,
	userID string,
) ([]Household, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}

	return s.repository.FindByUserID(ctx, uid)
}

func (s *Service) GetByID(
	ctx context.Context,
	householdID string,
	userID string,
) (Household, error) {
	hid, uid, err := parseIDs(householdID, userID)
	if err != nil {
		return Household{}, err
	}

	isMember, err := s.repository.IsMember(ctx, hid, uid)
	if err != nil {
		return Household{}, err
	}
	if !isMember {
		return Household{}, ErrHouseholdNotFound
	}

	household, err := s.repository.FindByID(ctx, hid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Household{}, ErrHouseholdNotFound
	}
	if err != nil {
		return Household{}, err
	}

	return household, nil
}

func (s *Service) ListMembers(
	ctx context.Context,
	householdID string,
	userID string,
) ([]MemberResponse, error) {
	hid, uid, err := parseIDs(householdID, userID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.repository.IsMember(ctx, hid, uid)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrHouseholdNotFound
	}

	// Pastikan household memang ada.
	_, err = s.repository.FindByID(ctx, hid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrHouseholdNotFound
	}
	if err != nil {
		return nil, err
	}

	return s.repository.FindMembers(ctx, hid)
}

func parseIDs(
	householdID string,
	userID string,
) (bson.ObjectID, bson.ObjectID, error) {
	hid, err := bson.ObjectIDFromHex(householdID)
	if err != nil {
		return bson.NilObjectID, bson.NilObjectID, ErrInvalidID
	}

	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return bson.NilObjectID, bson.NilObjectID, ErrInvalidID
	}

	return hid, uid, nil
}
