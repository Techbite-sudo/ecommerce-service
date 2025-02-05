package middleware

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidGrant = errors.New("invalid grant")
	ErrTokenExpired = errors.New("token expired")
)

// GetAuthCodeURL generates the authorization URL for OIDC login
func GetAuthCodeURL(state string, redirectURI string) string {
	oauth2Config.RedirectURL = redirectURI
	return oauth2Config.AuthCodeURL(state)
}

// ExchangeCodeForToken exchanges the authorization code for an ID token
func ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	// Exchange code for token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", ErrInvalidGrant
	}

	// Extract the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", errors.New("no id_token in token response")
	}

	// Verify the ID token
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	// Verify expiration
	if idToken.Expiry.Before(time.Now()) {
		return "", ErrTokenExpired
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return "", err
	}

	// Create or update user in database
	_, err = getOrCreateUser(claims)
	if err != nil {
		return "", err
	}

	// Return the ID token
	return rawIDToken, nil
}

// ValidateToken validates an ID token and returns the claims
func ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	idToken, err := verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}
