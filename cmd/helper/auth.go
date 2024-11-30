package helper

import (
	"crypto"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	key crypto.PublicKey
}

func NewValidator(path string) (*Validator, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key file: %w", err)
	}

	key, err := jwt.ParseEdPublicKeyFromPEM(bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse as ed public key: %w", err)
	}

	return &Validator{key: key}, nil
}

func (v *Validator) GetToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return v.key, nil
		})
	if err != nil {
		return nil, fmt.Errorf("unable to parse token string: %w", err)
	}

	aud, ok := token.Claims.(jwt.MapClaims)["aud"]
	if !ok {
		return nil, fmt.Errorf("token had no audience claim")
	}

	if aud != "api" {
		return nil, fmt.Errorf("token had the wrong audience claim")
	}

	return token, nil
}
