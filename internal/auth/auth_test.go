package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserId  uuid.UUID
		wantErr     bool
	}{
		{
			"Valid token",
			validToken,
			"secret",
			userID,
			false,
		},
		{
			"Invalid token",
			"invalid-token-string",
			"secret",
			uuid.Nil,
			true,
		},
		{
			"Wrong secret",
			validToken,
			"wrong_secret",
			uuid.Nil,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userId, err := ValidateJWT(test.tokenString, test.tokenSecret)
			if (err != nil) != test.wantErr {
				t.Errorf("ValidateJWT()\nerror = %v\nwantErr = %v", test.wantErr, err)
				return
			}

			if userId != test.wantUserId {
				t.Errorf("ValidateJWT()\nuserId = %v\nwantUserId = %v", userId, test.wantUserId)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			"Valid header",
			http.Header{
				"Authorization": []string{"Bearer token"},
			},
			"token",
			false,
		},
		{
			"Missing Authorization header",
			http.Header{},
			"",
			true,
		},
		{
			"Malformed Authorization header",
			http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			"",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := GetBearerToken(test.headers)
			if (err != nil) != test.wantErr {
				t.Errorf("GetBearerToken()\nerror = %v\nwantErr = %v", test.wantErr, err)
				return
			}

			if token != test.wantToken {
				t.Errorf("GetBearerToken()\ntoken = %v\nwantToken = %v", token, test.wantToken)
			}
		})
	}
}
