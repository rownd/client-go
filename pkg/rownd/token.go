package rownd

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type AuthLevel string

const (
	AuthLevelInstant    AuthLevel = "instant"
	AuthLevelUnverified AuthLevel = "unverified"
	AuthLevelGuest      AuthLevel = "guest"
	AuthLevelVerified   AuthLevel = "verified"
)

// Token ...
type Token struct {
	*jwt.Token

	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token"`
	Claims      Claims `json:"decoded_token"`
}

// Claims ...
type Claims struct {
	Exp *jwt.NumericDate `json:"exp"`
	Sub string           `json:"sub"`
	Iss string           `json:"iss"`
	Aud jwt.ClaimStrings `json:"aud"`
	Iat *jwt.NumericDate `json:"iat"`
	Nbf *jwt.NumericDate `json:"nbf"`
	Jti string           `json:"jti"`

	// custom rownd claims
	AppUserID      string    `json:"https://auth.rownd.io/app_user_id"`
	IsUserVerified bool      `json:"https://auth.rownd.io/is_verified_user"`
	IsAnonymous    bool      `json:"https://auth.rownd.io/is_anonymous"`
	AuthLevel      AuthLevel `json:"https://auth.rownd.io/auth_level"`
}

// GetExpirationTime implements interface method jwt.Claims.GetExpirationTime()
func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.Exp, nil
}

// GetIssuedAt implements interface method jwt.Claims.GetIssuedAt()
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.Iat, nil
}

// GetNotBefore implements interface method jwt.Claims.GetNotBefore()
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.Nbf, nil
}

// GetIssuer implements interface method jwt.Claims.GetIssuer()
func (c Claims) GetIssuer() (string, error) {
	return c.Iss, nil
}

// GetSubject implements interface method jwt.Claims.GetSubject()
func (c Claims) GetSubject() (string, error) {
	return c.Sub, nil
}

// GetAudience implements interface method jwt.Claims.GetAudience()
func (c Claims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Aud, nil
}

// rowndCtxKey ...
type rowndCtxKey struct{}

// AddTokenToCtx embeds the token in the request context.
func AddTokenToCtx(ctx context.Context, value *Token) context.Context {
	return context.WithValue(ctx, rowndCtxKey{}, value)
}

// TokenFromCtx extracts the token embedded in the request context.
func TokenFromCtx(ctx context.Context) *Token {
	res := ctx.Value(rowndCtxKey{})
	if res == nil {
		return nil
	}

	t, ok := res.(*Token)
	if !ok {
		return nil
	}

	return t
}

// TokenValidator ...
type TokenValidator interface {
	Validate(ctx context.Context, token string) (*Token, error)
}

type tokenValidator struct {
	*Client
}

// Validate ...
func (c *tokenValidator) Validate(ctx context.Context, token string) (*Token, error) {
	if token == "" {
		return nil, NewError(ErrAuthentication, "invalid token", nil)
	}

	// fetch JWKS
	jwks, err := c.fetchJWKS(ctx)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to fetch JWKS", err)
	}

	// parse and validate the token
	parsedToken, err := jwt.ParseWithClaims(token, Claims{}, func(token *jwt.Token) (any, error) {
		// FIXME does rownd support other signing alg?
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		// find the key with matching kid
		key, ok := jwks.Contains(kid)
		if !ok {
			return nil, fmt.Errorf("key %s not found", kid)
		}

		// decode the EdDSA public key from base64
		publicKey, derr := base64.RawURLEncoding.DecodeString(key.X)
		if derr != nil {
			return nil, fmt.Errorf("invalid key format for key %s: %w", key.KID, err)
		}

		return ed25519.PublicKey(publicKey), nil
	})
	if err != nil {
		return nil, NewError(ErrAuthentication, "invalid token", err)
	}

	if !parsedToken.Valid {
		return nil, NewError(ErrAuthentication, "invalid token", nil)
	}

	claims, ok := parsedToken.Claims.(Claims)
	if !ok || !parsedToken.Valid {
		return nil, NewError(ErrAuthentication, "invalid token claims", nil)
	}

	r := Token{
		Claims:      claims,
		UserID:      claims.AppUserID,
		AccessToken: token,
	}

	return &r, nil
}
