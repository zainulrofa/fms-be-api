package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

const accessTokenTTL = 15 * time.Minute

type RegisterInput struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	HouseholdName string `json:"householdName"`
}

type RegisterResult struct {
	UserID      bson.ObjectID `json:"userId"`
	HouseholdID bson.ObjectID `json:"householdId"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	UserID      bson.ObjectID `json:"userId"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	AccessToken string        `json:"accessToken"`
	TokenType   string        `json:"tokenType"`
	ExpiresIn   int64         `json:"expiresIn"`
	ExpiresAt   time.Time     `json:"expiresAt"`
}

type ProfileResult struct {
	UserID    bson.ObjectID `json:"userId"`
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	AvatarURL string        `json:"avatarUrl,omitempty"`
	IsActive  bool          `json:"isActive"`
	CreatedAt time.Time     `json:"createdAt"`
}

type Service struct {
	repository *Repository
	jwtSecret  []byte
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func NewServiceWithJWT(repository *Repository, secret string) *Service {
	return &Service{
		repository: repository,
		jwtSecret:  []byte(secret),
	}
}

func (s *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (RegisterResult, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	householdName := strings.TrimSpace(input.HouseholdName)

	if name == "" || householdName == "" {
		return RegisterResult{}, ErrInvalidInput
	}

	if !utf8.ValidString(name) ||
		utf8.RuneCountInString(name) > 100 ||
		!utf8.ValidString(householdName) ||
		utf8.RuneCountInString(householdName) > 100 {
		return RegisterResult{}, ErrInvalidInput
	}

	if err := validateEmail(email); err != nil {
		return RegisterResult{}, err
	}

	if len(input.Password) < 8 ||
		len(input.Password) > 72 ||
		!utf8.ValidString(input.Password) {
		return RegisterResult{}, fmt.Errorf(
			"%w: password must contain 8 to 72 bytes",
			ErrInvalidInput,
		)
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	userID := bson.NewObjectID()
	householdID := bson.NewObjectID()

	user := User{
		ID:           userID,
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	household := Household{
		ID:        householdID,
		Name:      householdName,
		OwnerID:   userID,
		Currency:  "IDR",
		Timezone:  "Asia/Jakarta",
		CreatedAt: now,
		UpdatedAt: now,
	}

	member := HouseholdMember{
		ID:          bson.NewObjectID(),
		HouseholdID: householdID,
		UserID:      userID,
		Role:        "OWNER",
		JoinedAt:    now,
		CreatedAt:   now,
	}

	if err := s.repository.Register(ctx, user, household, member); err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		UserID:      userID,
		HouseholdID: householdID,
		Name:        name,
		Email:       email,
	}, nil
}

func (s *Service) Login(
	ctx context.Context,
	input LoginInput,
) (LoginResult, error) {
	if len(s.jwtSecret) < 32 {
		return LoginResult{}, errors.New("JWT secret is not configured correctly")
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.Password == "" {
		return LoginResult{}, ErrInvalidInput
	}

	if err := validateEmail(email); err != nil {
		return LoginResult{}, err
	}

	user, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !user.IsActive {
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	expiresAt := now.Add(accessTokenTTL)

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.Hex(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return LoginResult{}, fmt.Errorf("sign access token: %w", err)
	}

	return LoginResult{
		UserID:      user.ID,
		Name:        user.Name,
		Email:       user.Email,
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(accessTokenTTL.Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) GetProfile(
	ctx context.Context,
	userID string,
) (ProfileResult, error) {
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return ProfileResult{}, ErrInvalidInput
	}

	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return ProfileResult{}, err
	}

	if !user.IsActive {
		return ProfileResult{}, ErrUserNotFound
	}

	return ProfileResult{
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}, nil
}

func validateEmail(email string) error {
	if len(email) > 254 {
		return ErrInvalidEmail
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return ErrInvalidEmail
	}

	return nil
}
