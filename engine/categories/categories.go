package categories

import (
	"ecommerce-service/graph/model"
	"ecommerce-service/models"
	"ecommerce-service/utils"
	"errors"

	uuid "github.com/satori/go.uuid"
)

func CreateCategory(input model.CategoryInput) (*model.Category, error) {
	category := models.Category{
		Name: input.Name,
	}

	if input.ParentID != nil {
		parentUUID, err := uuid.FromString(*input.ParentID)
		if err != nil {
			return nil, err
		}

		// Verify parent exists and check level
		var parent models.Category
		if err := utils.DB.First(&parent, "id = ?", parentUUID).Error; err != nil {
			return nil, err
		}

		// Prevent deep nesting (more than 5 levels)
		if parent.Level >= 4 {
			return nil, errors.New("maximum category nesting level reached")
		}

		category.ParentID = &parentUUID
	}

	if err := utils.DB.Create(&category).Error; err != nil {
		return nil, err
	}

	return category.ToGraphQL(), nil
}

func GetCategories() ([]*model.Category, error) {
	var categories []models.Category
	if err := utils.DB.Preload("Children").Preload("Products").Find(&categories).Error; err != nil {
		return nil, err
	}

	result := make([]*model.Category, len(categories))
	for i, category := range categories {
		result[i] = category.ToGraphQL()
	}

	return result, nil
}

func GetCategoryByID(id string) (*model.Category, error) {
	catUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	var category models.Category
	if err := utils.DB.Preload("Children").Preload("Products").First(&category, "id = ?", catUUID).Error; err != nil {
		return nil, err
	}

	return category.ToGraphQL(), nil
}

func UpdateCategory(id string, input model.CategoryInput) (*model.Category, error) {
	catUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	var category models.Category
	if err := utils.DB.First(&category, "id = ?", catUUID).Error; err != nil {
		return nil, err
	}

	// Update the name
	category.Name = input.Name

	// Update parent if provided
	if input.ParentID != nil {
		parentUUID, err := uuid.FromString(*input.ParentID)
		if err != nil {
			return nil, err
		}

		// Verify parent exists
		var parent models.Category
		if err := utils.DB.First(&parent, "id = ?", parentUUID).Error; err != nil {
			return nil, err
		}

		// Check for circular reference
		if parent.ID == category.ID {
			return nil, errors.New("category cannot be its own parent")
		}

		// Check if new parent would create too deep nesting
		if parent.Level >= 4 {
			return nil, errors.New("maximum category nesting level would be exceeded")
		}

		category.ParentID = &parentUUID
		category.Level = parent.Level + 1
	} else {
		category.ParentID = nil
		category.Level = 0
	}

	// Save updates
	if err := utils.DB.Save(&category).Error; err != nil {
		return nil, err
	}

	// Reload category with relationships
	if err := utils.DB.Preload("Children").Preload("Products").First(&category, "id = ?", catUUID).Error; err != nil {
		return nil, err
	}

	return category.ToGraphQL(), nil
}
