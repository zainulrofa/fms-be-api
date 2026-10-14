package households

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const CollectionName = "households"

type Household struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	OwnerID   bson.ObjectID `bson:"ownerId" json:"ownerId"`
	Currency  string        `bson:"currency" json:"currency"`
	Timezone  string        `bson:"timezone" json:"timezone"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}
