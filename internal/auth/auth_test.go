package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	tests := []struct {
		name      string
		userId    uuid.UUID
		secret    string
		expiresIn time.Duration
		want      uuid.UUID
		wantErr   bool
	}{
		{"success jwt validation", id1, "test123", 100 * time.Second, id1, false},
		{"wrong jwt validation", id2, "thisIsASecret", 1 * time.Second, uuid.Nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := MakeJWT(test.userId, test.secret, test.expiresIn)
			if err != nil {
				t.Errorf("unexpected error got %v", err)
			}

			time.Sleep(2 * time.Second)

			id, err := ValidateJWT(token, test.secret)
			if (err != nil) != test.wantErr {
				t.Errorf("expected error %v, got %v", test.wantErr, err)
			}

			if id != test.want {
				t.Errorf("expected %v, got %v", test.want, id)
			}
		})
	}
}
