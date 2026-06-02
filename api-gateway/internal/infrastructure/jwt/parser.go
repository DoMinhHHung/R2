package jwt

import (
	"errors"
	"strings"

	"github.com/DoMinhHHung/Rental/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
)

type Parser struct {
	secret []byte
	issuer string
}

func NewParser(secret, issuer string) *Parser {
	return &Parser{secret: []byte(secret), issuer: issuer}
}

type customClaims struct {
	UserID string   `json:"uid"`
	Roles  []string `json:"roles"`
	APIKey string   `json:"api_key"`
	jwt.RegisteredClaims
}

func (p *Parser) Parse(tokenStr string) (*entity.Claims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims")
	}

	if claims.Issuer != p.issuer {
		return nil, errors.New("invalid issuer")
	}

	return &entity.Claims{
		UserID: claims.UserID,
		Roles:  claims.Roles,
		APIKey: claims.APIKey,
	}, nil
}
