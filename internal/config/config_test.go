package config

import "testing"

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		env       string
		mongoURI  string
		wantPort  string
		wantEnv   string
		wantError bool
	}{
		{
			name:     "valid configuration",
			port:     "8081",
			env:      "development",
			mongoURI: "mongodb://localhost:27017/fms",
			wantPort: "8081",
			wantEnv:  "development",
		},
		{
			name:     "default port and environment",
			mongoURI: "mongodb://localhost:27017/fms",
			wantPort: "8080",
			wantEnv:  "development",
		},
		{
			name:      "missing MongoDB URI",
			env:       "development",
			wantError: true,
		},
		{
			name:      "invalid environment",
			env:       "staging",
			mongoURI:  "mongodb://localhost:27017/fms",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_PORT", tt.port)
			t.Setenv("APP_ENV", tt.env)
			t.Setenv("MONGODB_URI", tt.mongoURI)
			t.Setenv(
				"JWT_ACCESS_SECRET",
				"test-secret-for-fms-unit-tests-123456",
			)
			got, err := Load()

			if tt.wantError {
				if err == nil {
					t.Fatal("Load() expected an error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}

			if got.AppPort != tt.wantPort {
				t.Errorf("AppPort = %q, want %q", got.AppPort, tt.wantPort)
			}

			if got.AppEnv != tt.wantEnv {
				t.Errorf("AppEnv = %q, want %q", got.AppEnv, tt.wantEnv)
			}
		})
	}
}
