package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var tokenClaimError = errors.New("unable to parse JWT: unexpected claims type")

// Token wraps a raw JWT and its claims.
type Token struct {
	Raw    string
	claims *jwt.StandardClaims
}

// NewToken initializes a new Token object by parsing the claims of the provided raw JWT.
func NewToken(raw string) (*Token, error) {
	claims, err := parseClaims(raw)
	if err != nil {
		return nil, err
	}

	return &Token{
		Raw:    raw,
		claims: claims,
	}, nil
}

// IsExpired checks whether the token has expired.
func (t *Token) IsExpired() bool {
	required := true

	return !t.claims.VerifyExpiresAt(time.Now().Unix(), required)
}

// parseClaims returns the JWT claims without checking the header or signature.
func parseClaims(raw string) (*jwt.StandardClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(raw, &jwt.StandardClaims{})
	if err != nil {
		return nil, fmt.Errorf("unable to parse JWT: %s", err)
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims == nil {
		return nil, tokenClaimError
	}

	return claims, nil
}
