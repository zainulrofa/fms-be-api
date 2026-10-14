package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		unauthorized := func() {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"status": "error",
					"error": gin.H{
						"code":    "unauthorized",
						"message": "A valid access token is required",
					},
				},
			)
		}

		if len(secret) < 32 {
			unauthorized()
			return
		}

		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			unauthorized()
			return
		}

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)

		if err != nil || token == nil || !token.Valid {
			unauthorized()
			return
		}

		// Pastikan subject merupakan MongoDB ObjectID yang valid.
		if _, err := bson.ObjectIDFromHex(claims.Subject); err != nil {
			unauthorized()
			return
		}

		c.Set("authenticatedUserID", claims.Subject)
		c.Next()
	}
}
