package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTPayload represents the JWT token payload
type JWTPayload struct {
	UserID        string   `json:"userId"`
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	CompanyID     string   `json:"companyId"`
	Role          string   `json:"role"`
	CompanyAccess []string `json:"companyAccess"`
	Permissions   []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWT utility for token generation and validation
type JWTUtil struct {
	secret    []byte
	expiresIn time.Duration
}

// NewJWTUtil creates a new JWT utility
func NewJWTUtil(secret string, expiresIn time.Duration) *JWTUtil {
	return &JWTUtil{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

// GenerateToken creates a new JWT token
func (j *JWTUtil) GenerateToken(payload JWTPayload) (string, error) {
	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    "pos-retail",
		Subject:   payload.UserID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":        payload.UserID,
		"username":      payload.Username,
		"email":         payload.Email,
		"companyId":     payload.CompanyID,
		"role":          payload.Role,
		"companyAccess": payload.CompanyAccess,
		"permissions":   payload.Permissions,
		"jti":           claims.ID,
		"iss":           claims.Issuer,
		"sub":           claims.Subject,
		"iat":           claims.IssuedAt,
		"exp":           claims.ExpiresAt,
	})

	return token.SignedString(j.secret)
}

// VerifyToken validates a JWT token
func (j *JWTUtil) VerifyToken(tokenString string) (*JWTPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	payload := &JWTPayload{
		UserID:        getStringClaim(claims, "userId"),
		Username:      getStringClaim(claims, "username"),
		Email:         getStringClaim(claims, "email"),
		CompanyID:     getStringClaim(claims, "companyId"),
		Role:          getStringClaim(claims, "role"),
		CompanyAccess: getStringSliceClaim(claims, "companyAccess"),
		Permissions:   getStringSliceClaim(claims, "permissions"),
	}

	return payload, nil
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSliceClaim(claims jwt.MapClaims, key string) []string {
	if val, ok := claims[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, v := range slice {
				if str, ok := v.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return nil
}

// GetTokenExpiry returns the token expiry time
func (j *JWTUtil) GetTokenExpiry() time.Time {
	return time.Now().Add(j.expiresIn)
}
