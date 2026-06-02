package jwt

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/DoMinhHHung/R2/internal/domain/entity"
	gojwt "github.com/golang-jwt/jwt/v5"
)

type Parser struct {
	publicKey any // *rsa.PublicKey
	issuer    string
}

func NewParser(publicKeyPath, issuer string) (*Parser, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key %q: %w", publicKeyPath, err)
	}

	pubKey, err := gojwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA public key: %w", err)
	}

	return &Parser{publicKey: pubKey, issuer: issuer}, nil
}

type claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

func (p *Parser) Parse(tokenStr string) (*entity.Claims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}

	token, err := gojwt.ParseWithClaims(tokenStr, &claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.publicKey, nil
	}, gojwt.WithIssuer(p.issuer))

	if err != nil {
		return nil, err
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims")
	}

	return &entity.Claims{
		UserID: c.UserID,
		Roles:  []string{c.Role},
	}, nil
}
