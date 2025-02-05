package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	// "time"
	"github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
)

// Generate an access token for the authenticated user
func GenerateAccessToken(userID uuid.UUID) (string, error) {
	privateKey, err := loadRSAPrivateKey()
	if err != nil {
		return "", err
	}
	// Create a JWT token
	token, err := generateJWT(privateKey, userID)
	if err != nil {
		return "", nil
	}
	return token, nil
}

func loadRSAPrivateKey() (*rsa.PrivateKey, error) {
	keystring := os.Getenv("JWT_KEY")
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keystring))
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func generateJWT(privateKey *rsa.PrivateKey, authid uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = authid
	// claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func loadX509PublicKey() (*rsa.PublicKey, error) {
	certstring := os.Getenv("JWT_CERT")
	block, _ := pem.Decode([]byte(certstring))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate contains a non-RSA public key")
	}

	return publicKey, nil
}

func VerifyAccessToken(tokenString string) (string, error) {
	publicKey, err := loadX509PublicKey()
	if err != nil {
		return "", err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}
	return "nil", errors.New("jwt verification failed")
}
