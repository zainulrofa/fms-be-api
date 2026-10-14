package users

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const CollectionName = "users"

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"passwordHash" json:"-"`
	AvatarURL    string        `bson:"avatarUrl,omitempty" json:"avatarUrl,omitempty"`
	IsActive     bool          `bson:"isActive" json:"isActive"`
	LastLoginAt  *time.Time    `bson:"lastLoginAt,omitempty" json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt" json:"updatedAt"`
}
