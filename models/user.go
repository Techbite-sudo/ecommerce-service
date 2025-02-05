package models

import (
	"ecommerce-service/graph/model"
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Base
	Names               string `gorm:"not null"`
	Email               string `gorm:"not null;uniqueIndex"`
	Password            string `gorm:"not null"`
	PhoneNumber         string `gorm:"not null"`
	Country             string `gorm:"not null"`
	Role                Role   `gorm:"not null;type:text"`
	LastLoginAt         *time.Time
	PasswordResetToken  string
	PasswordResetExpiry *time.Time
	Orders              []Order `gorm:"foreignKey:CustomerID"`
}

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) ToGraphData() *model.User {
	return &model.User{
		ID:          u.ID.String(),
		Names:       u.Names,
		Email:       u.Email,
		Password:    u.Password,
		PhoneNumber: u.PhoneNumber,
		Country:     u.Country,
		Role:        model.Role(u.Role),
		CreatedAt:   u.CreatedAt,
	}
}

func (u *User) GeneratePasswordResetToken() (string, error) {
	// Generate a UUID for the reset token
	resetToken := uuid.NewV4().String()

	// Set token and expiry (24 hours from now)
	u.PasswordResetToken = resetToken
	expiry := time.Now().Add(24 * time.Hour)
	u.PasswordResetExpiry = &expiry

	return resetToken, nil
}

func (u *User) IsPasswordResetTokenValid(token string) bool {
	if u.PasswordResetToken != token {
		return false
	}

	if u.PasswordResetExpiry == nil || u.PasswordResetExpiry.Before(time.Now()) {
		return false
	}

	return true
}

func (u *User) ClearPasswordResetToken() {
	u.PasswordResetToken = ""
	u.PasswordResetExpiry = nil
}

func (u *User) UpdateLoginTimestamp() {
	now := time.Now()
	u.LastLoginAt = &now
}
