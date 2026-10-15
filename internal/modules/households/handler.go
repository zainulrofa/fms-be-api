package households

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}

	households, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list households", "error", err)
		writeError(c, http.StatusInternalServerError, "internal_error", "Failed to retrieve households")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   households,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}

	household, err := h.service.GetByID(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handleHouseholdError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   household,
	})
}

func (h *Handler) ListMembers(c *gin.Context) {
	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}

	members, err := h.service.ListMembers(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handleHouseholdError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   members,
	})
}

func getAuthenticatedUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get("authenticatedUserID")
	if !exists {
		return "", false
	}

	switch id := value.(type) {
	case string:
		return id, id != ""
	case bson.ObjectID:
		return id.Hex(), true
	default:
		return "", false
	}
}

func handleHouseholdError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidID):
		writeError(c, http.StatusBadRequest, "invalid_id", "Household ID is invalid")

	case errors.Is(err, ErrHouseholdNotFound):
		// Household yang tidak ada dan household yang bukan milik user
		// menggunakan respons yang sama.
		writeError(c, http.StatusNotFound, "household_not_found", "Household not found")

	default:
		slog.Error("household request failed", "error", err)
		writeError(c, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"status": "error",
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
