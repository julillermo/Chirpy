package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ? JULIUS: This is honestly a little rough. I made this just enough to get the base idea
// I think it would be best to learn about testing in go.
// Expecting tests in go to be a lot like testing in the JS/TS ecosystem.
func TestMakeJWT(t *testing.T) {
	type makeJWTStruct struct {
		userId      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}

	expectedUUID := uuid.New()
	wrongUUID := uuid.New()

	targetJWT := makeJWTStruct{
		userId:      expectedUUID,
		tokenSecret: "Totes-very_special=$ecret",
		expiresIn:   time.Hour,
	}

	cases := []struct {
		name      string
		input     makeJWTStruct
		expectErr bool
	}{
		{
			name:      "Valid token",
			input:     targetJWT,
			expectErr: false,
		},
		{
			name: "Invalid token",
			input: makeJWTStruct{
				userId:      wrongUUID,
				tokenSecret: "Totes-very_special=$ecret",
				expiresIn:   time.Duration(time.Duration.Hours(1)),
			},
			expectErr: true,
		},
		{
			name: "Wrong secret",
			input: makeJWTStruct{
				userId:      expectedUUID,
				tokenSecret: "notVerySecret",
				expiresIn:   time.Duration(time.Duration.Hours(1)),
			},
			expectErr: true,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			newJWT, newJWTErr := MakeJWT(
				targetJWT.userId,
				targetJWT.tokenSecret,
				targetJWT.expiresIn,
			)
			if newJWTErr != nil {
				t.Fatalf("MakeJWT() error = %v", newJWTErr)
			}

			validatedJWT, ValidateJWTErr := ValidateJWT(newJWT, cs.input.tokenSecret)
			if ValidateJWTErr != nil {
				if !cs.expectErr {
					t.Fatalf("ValidateJWT() error = %v", ValidateJWTErr)
				} else {
					return
				}
			}

			if validatedJWT != cs.input.userId {
				if !cs.expectErr {
					t.Fatalf("ValidateJWT() mismatch token error = %v =/= %v ", validatedJWT, cs.input.userId)
				} else {
					return
				}
			}

		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name                     string
		inputAuthorizationHeader string
		expectedToken            string
		expectFail               bool
	}{
		{
			name:                     "success",
			inputAuthorizationHeader: "Bearer hehehe",
			expectedToken:            "hehehe",
			expectFail:               false,
		},
		{
			name:                     "failure",
			inputAuthorizationHeader: "Bearer hehehe",
			expectedToken:            "hehehe",
			expectFail:               true,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Set("Authorization", cs.inputAuthorizationHeader)

			token, tokenErr := GetBearerToken(headers)

			if tokenErr != nil {
				t.Errorf("GetBearerToken(headers) error = %v", tokenErr)
				return
			}

			if token != cs.expectedToken {
				if !cs.expectFail {
					t.Errorf("GetBearerToken(headers) token mismatch = %v =/= %v", token, "hehehe")
				}
				return
			}
		})
	}
}
