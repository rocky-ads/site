package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/models"
)

// JWTMiddleware is a middleware that validates a JWT token and sets the user in the context.
func JWTMiddleware(c *fiber.Ctx) error {
	// Get JWT token from cookie
	tokenString := cookie.GetJWT(c)
	if tokenString == "" {
		local.SetUserID(c, 0)
		local.SetUserName(c, "")
		return c.Next()
	}

	// Validate JWT token
	claims, err := validateToken(tokenString)
	if err != nil {
		// Invalid token, clear cookie
		cookie.ClearJWT(c)
		local.SetUserID(c, 0)
		local.SetUserName(c, "")
		return c.Next()
	}

	// Set user ID and username in context
	local.SetUserID(c, getUserID(claims))
	local.SetUserName(c, getUserName(claims))
	return c.Next()
}

type claims struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

// generateToken creates a JWT token for a user
func generateToken(u *models.User) (string, error) {
	claims := claims{
		UserID:   u.ID,
		UserName: u.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   string(rune(u.ID)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

// validateToken validates a JWT token and returns the claims
func validateToken(tokenString string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// getUserID extracts the user ID from validated claims
func getUserID(claims *claims) int {
	return claims.UserID
}

// getUserName extracts the username from validated claims
func getUserName(claims *claims) string {
	return claims.UserName
}

// validateSecret validates that JWT secret is set and sufficiently strong
// Minimum requirements: at least 32 bytes (256 bits) of entropy
func validateSecret(secret []byte) error {
	if len(secret) == 0 {
		return fmt.Errorf("JWT_SECRET environment variable is required but not set")
	}

	// Require at least 32 bytes (256 bits) for HS256
	minLength := 32
	if len(secret) < minLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters long for security", minLength)
	}

	return nil
}
