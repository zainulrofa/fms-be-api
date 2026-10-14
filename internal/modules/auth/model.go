package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Name         string        `bson:"name"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"passwordHash"`
	AvatarURL    string        `bson:"avatarUrl,omitempty"`
	IsActive     bool          `bson:"isActive"`
	CreatedAt    time.Time     `bson:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt"`
}

type Household struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	OwnerID   bson.ObjectID `bson:"ownerId"`
	Currency  string        `bson:"currency"`
	Timezone  string        `bson:"timezone"`
	CreatedAt time.Time     `bson:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

type HouseholdMember struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	HouseholdID bson.ObjectID `bson:"householdId"`
	UserID      bson.ObjectID `bson:"userId"`
	Role        string        `bson:"role"`
	JoinedAt    time.Time     `bson:"joinedAt"`
	CreatedAt   time.Time     `bson:"createdAt"`
}
