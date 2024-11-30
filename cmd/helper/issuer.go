package helper

import (
	"crypto"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	key crypto.PrivateKey
}

func NewIssuer(path string) (*Issuer, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("unable to read private key file: %w", err))
	}

	key, err := jwt.ParseEdPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("unable to parse as ed private key: %w", err)
	}

	return &Issuer{key: key}, nil
}

func (i *Issuer) IssueToken(id uint, name string, roles []string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"aud": "api",
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 30).Unix(),
		"iss": "http://localhost:5000",
		//TODO convert uint to string
		"sub":   fmt.Sprint(id),
		"name":  name,
		"roles": roles,
	})

	tokenString, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	return tokenString, nil
}
