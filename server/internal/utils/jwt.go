package utils

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/ayussh-2/timepad/config"
	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTUtil struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTUtil(cfg *config.Config) (*JWTUtil, error) {
	privateBytes, err := os.ReadFile(cfg.JWTPrivateKey)
	if err != nil {
		return nil, errors.New("could not read private key file: " + err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		return nil, errors.New("could not parse private key: " + err.Error())
	}

	publicBytes, err := os.ReadFile(cfg.JWTPublicKey)
	if err != nil {
		return nil, errors.New("could not read public key file: " + err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		return nil, errors.New("could not parse public key: " + err.Error())
	}

	return &JWTUtil{
		privateKey:    privateKey,
		publicKey:     publicKey,
		accessExpiry:  time.Duration(cfg.JWTAccessExpiry) * time.Second,
		refreshExpiry: time.Duration(cfg.JWTRefreshExpiry) * time.Second,
	}, nil
}

// NewJWTUtilFromKeys creates a JWTUtil directly from in-memory RSA keys.
// Useful in tests to avoid reading key files from disk.
func NewJWTUtilFromKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *JWTUtil {
	return &JWTUtil{
		privateKey:    privateKey,
		publicKey:     publicKey,
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (j *JWTUtil) GenerateAccessToken(userID string) (string, error) {
	claims := AuthClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "timepad-server",
		},
	}
	// Sign with Private Key (RS256)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

func (j *JWTUtil) GenerateRefreshToken(userID string) (string, error) {
	claims := AuthClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "timepad-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

func (j *JWTUtil) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
