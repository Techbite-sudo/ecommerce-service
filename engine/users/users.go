package users

import (
	"ecommerce-service/engine/notifications"
	"ecommerce-service/graph/model"
	"ecommerce-service/models"
	"ecommerce-service/utils"
	"errors"

	uuid "github.com/satori/go.uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already registered")
)

func FetchUserByID(id string) (*model.User, error) {
	userID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := utils.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	return user.ToGraphData(), nil
}

func FetchUser(email string) (*model.User, error) {
	var user models.User
	if err := utils.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	return user.ToGraphData(), nil
}

func PasswordResetRequest(email string) (string, error) {
	var user models.User
	if err := utils.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", ErrUserNotFound
	}

	// Generate reset token
	resetToken, err := user.GeneratePasswordResetToken()
	if err != nil {
		return "", err
	}

	// Save user with reset token
	if err := utils.DB.Save(&user).Error; err != nil {
		return "", err
	}

	// Send password reset email
	if err := notifications.SendPasswordResetEmail(user.Email, resetToken); err != nil {
		return "", err
	}

	return "Password reset email sent successfully", nil
}

func ResetPassword(input *model.PasswordResetInput) (bool, error) {
	if input.NewPassword != input.ConfirmPassword {
		return false, errors.New("passwords do not match")
	}

	var user models.User
	if err := utils.DB.Where("password_reset_token = ?", input.Token).First(&user).Error; err != nil {
		return false, errors.New("invalid reset token")
	}

	if !user.IsPasswordResetTokenValid(input.Token) {
		return false, errors.New("reset token has expired")
	}

	if err := user.SetPassword(input.NewPassword); err != nil {
		return false, err
	}

	user.ClearPasswordResetToken()
	if err := utils.DB.Save(&user).Error; err != nil {
		return false, err
	}

	return true, nil
}

func UpdateUserProfile(userID string, input model.UpdateProfileInput) (*model.User, error) {
	var user models.User
	if err := utils.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if input.PhoneNumber != nil {
		user.PhoneNumber = *input.PhoneNumber
	}
	if input.Country != nil {
		user.Country = *input.Country
	}

	if err := utils.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return user.ToGraphData(), nil
}
