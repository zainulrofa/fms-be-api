package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Status string   `json:"status"`
	Error  apiError `json:"error"`
}

type registerSuccessResponse struct {
	Status string         `json:"status"`
	Data   RegisterResult `json:"data"`
}

type loginSuccessResponse struct {
	Status string      `json:"status"`
	Data   LoginResult `json:"data"`
}

func (h *Handler) Register(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		1<<20,
	)

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"invalid_request",
			"Request body must contain valid JSON",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		10*time.Second,
	)
	defer cancel()

	result, err := h.service.Register(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			writeError(
				c,
				http.StatusBadRequest,
				"invalid_input",
				"Name, household name, or password is invalid",
			)

		case errors.Is(err, ErrInvalidEmail):
			writeError(
				c,
				http.StatusBadRequest,
				"invalid_email",
				"Please provide a valid email address",
			)

		case errors.Is(err, ErrEmailAlreadyExists):
			writeError(
				c,
				http.StatusConflict,
				"email_already_exists",
				"An account with this email already exists",
			)

		default:
			slog.Error("user registration failed", "error", err)
			writeError(
				c,
				http.StatusInternalServerError,
				"internal_server_error",
				"An unexpected error occurred",
			)
		}
		return
	}

	c.JSON(http.StatusCreated, registerSuccessResponse{
		Status: "success",
		Data:   result,
	})
}

func (h *Handler) Login(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		1<<20,
	)

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"invalid_request",
			"Request body must contain valid JSON",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		10*time.Second,
	)
	defer cancel()

	result, err := h.service.Login(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			writeError(
				c,
				http.StatusBadRequest,
				"invalid_input",
				"Email and password are required",
			)

		case errors.Is(err, ErrInvalidEmail):
			writeError(
				c,
				http.StatusBadRequest,
				"invalid_email",
				"Please provide a valid email address",
			)

		case errors.Is(err, ErrInvalidCredentials):
			writeError(
				c,
				http.StatusUnauthorized,
				"invalid_credentials",
				"Email or password is incorrect",
			)

		default:
			slog.Error("user login failed", "error", err)
			writeError(
				c,
				http.StatusInternalServerError,
				"internal_server_error",
				"An unexpected error occurred",
			)
		}
		return
	}

	c.JSON(http.StatusOK, loginSuccessResponse{
		Status: "success",
		Data:   result,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("authenticatedUserID")
	if userID == "" {
		writeError(
			c,
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication is required",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		10*time.Second,
	)
	defer cancel()

	profile, err := h.service.GetProfile(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput),
			errors.Is(err, ErrUserNotFound):
			writeError(
				c,
				http.StatusUnauthorized,
				"unauthorized",
				"Authentication is invalid or the account is unavailable",
			)

		default:
			slog.Error("get user profile failed", "error", err)
			writeError(
				c,
				http.StatusInternalServerError,
				"internal_server_error",
				"An unexpected error occurred",
			)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profile,
	})
}

func writeError(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, errorResponse{
		Status: "error",
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}
