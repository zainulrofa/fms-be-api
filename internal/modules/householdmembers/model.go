package householdmembers

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const CollectionName = "household_members"

type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
)

type HouseholdMember struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	HouseholdID bson.ObjectID `bson:"householdId" json:"householdId"`
	UserID      bson.ObjectID `bson:"userId" json:"userId"`
	Role        Role          `bson:"role" json:"role"`
	JoinedAt    time.Time     `bson:"joinedAt" json:"joinedAt"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
}
