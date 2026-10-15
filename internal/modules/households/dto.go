package households

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MemberResponse struct {
	UserID   bson.ObjectID `bson:"userId" json:"userId"`
	Role     string        `bson:"role" json:"role"`
	JoinedAt time.Time     `bson:"joinedAt" json:"joinedAt"`
}
