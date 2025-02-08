// internal/google_playstore/rtdn/validator/jwt_validator.go
package validator

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
)

// GoogleIssuer is the expected issuer for Play Store JWTs
const GoogleIssuer = "https://accounts.google.com"

var cachedKeys *idtoken.Payload

// ValidateJWT validates the webhook's Authorization header
func ValidateJWT(authHeader string, expectedAudience string) (*jwt.Token, error) {
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid Authorization header format")
	}
	tokenString := parts[1]

	// Use cached keys if available
	if cachedKeys != nil {
		return &jwt.Token{Valid: true}, nil
	}

	// Fetch Google's public keys for validation
	payload, err := idtoken.Validate(context.Background(), tokenString, expectedAudience)
	if err != nil {
		log.Println("❌ Invalid JWT:", err)
		return nil, errors.New("JWT validation failed")
	}

	// Cache the keys for future requests
	cachedKeys = payload

	log.Println("✅ Webhook JWT successfully validated")
	return nil, nil
}
