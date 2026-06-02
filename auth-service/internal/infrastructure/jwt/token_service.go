package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/DoMinhHHung/auth-service/internal/domain/entity"
	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	gojwt "github.com/golang-jwt/jwt/v5"
)

type service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessExp  time.Duration
	refreshExp time.Duration
	issuer     string
}

type jwtClaims struct {
	UserID    string `json:"uid"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	TokenType string `json:"type"`
	gojwt.RegisteredClaims
}

func New(cfg config.JWTConfig) (port.TokenService, error) {
	privateBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	privateKey, err := gojwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	publicBytes, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	publicKey, err := gojwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &service{
		privateKey: privateKey,
		publicKey:  publicKey,
		accessExp:  cfg.AccessTokenExp,
		refreshExp: cfg.RefreshTokenExp,
		issuer:     cfg.Issuer,
	}, nil
}

func (s *service) GenerateAccessToken(user *entity.AuthUser, sessionID string) (string, error) {
	return s.generate(user, sessionID, "access", s.accessExp)
}

func (s *service) GenerateRefreshToken(user *entity.AuthUser, sessionID string) (string, error) {
	return s.generate(user, sessionID, "refresh", s.refreshExp)
}

func (s *service) generate(user *entity.AuthUser, sessionID, tokenType string, exp time.Duration) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(exp)),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *service) ParseToken(tokenStr string) (*port.TokenClaims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	}, gojwt.WithIssuer(s.issuer))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &port.TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		SessionID: claims.SessionID,
		TokenType: claims.TokenType,
	}, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
