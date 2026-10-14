package auth

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   RegisterInput
		wantErr error
	}{
		{
			name: "empty name",
			input: RegisterInput{
				Email:         "user@example.com",
				Password:      "password123",
				HouseholdName: "My Household",
			},
			wantErr: ErrInvalidInput,
		},
		{
			name: "empty household name",
			input: RegisterInput{
				Name:     "Test User",
				Email:    "user@example.com",
				Password: "password123",
			},
			wantErr: ErrInvalidInput,
		},
		{
			name: "invalid email",
			input: RegisterInput{
				Name:          "Test User",
				Email:         "invalid-email",
				Password:      "password123",
				HouseholdName: "My Household",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "password too short",
			input: RegisterInput{
				Name:          "Test User",
				Email:         "user@example.com",
				Password:      "short",
				HouseholdName: "My Household",
			},
			wantErr: ErrInvalidInput,
		},
		{
			name: "password exceeds bcrypt limit",
			input: RegisterInput{
				Name:          "Test User",
				Email:         "user@example.com",
				Password:      string(make([]byte, 73)),
				HouseholdName: "My Household",
			},
			wantErr: ErrInvalidInput,
		},
	}

	service := NewService(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(
				context.Background(),
				tt.input,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}
