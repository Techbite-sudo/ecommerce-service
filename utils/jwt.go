package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const hashcost = 11

func HashString(str string) (string, error) {
	strByte := []byte(str)

	strHashed, err := bcrypt.GenerateFromPassword(strByte, hashcost)
	if err != nil {
		return "", err
	}
	return string(strHashed), nil
}

func CompareHashedString(hashedStr, str string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedStr), []byte(str))
	return err == nil
}

func GenerateJWTForAuthId(authId uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	signingKey := []byte(os.Getenv("JWT_SECRET"))
	claims := token.Claims.(jwt.MapClaims)
	claims["authId"] = authId.String()
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWTForAuthId(tokenString string) (string, error) {
	signingKey := []byte(os.Getenv("JWT_SECRET"))

	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if token == nil {
		return "", errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["authId"].(string), nil
	}

	return "", errors.New("invalid token")
}
