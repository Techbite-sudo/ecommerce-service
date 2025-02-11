package products

import (
	"ecommerce-service/graph/model"
	"ecommerce-service/models"
	"ecommerce-service/utils"

	uuid "github.com/satori/go.uuid"
)

func CreateProduct(input model.ProductInput) (*model.Product, error) {
	// Convert category IDs to UUIDs
	categories := make([]models.Category, 0)
	for _, categoryID := range input.CategoryIds {
		catUUID, err := uuid.FromString(categoryID)
		if err != nil {
			return nil, err
		}

		var category models.Category
		if err := utils.DB.First(&category, "id = ?", catUUID).Error; err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	// Create new product
	product := models.Product{
		Name:        input.Name,
		Description: *input.Description,
		Price:       input.Price,
		SKU:         input.Sku,
		Stock:       int(input.Stock),
		Categories:  categories,
	}

	if err := utils.DB.Create(&product).Error; err != nil {
		return nil, err
	}

	return product.ToGraphQL(), nil
}

func GetProducts(categoryID *string, search *string) ([]*model.Product, error) {
	var products []models.Product
	query := utils.DB.Preload("Categories")

	// Apply category filter if provided
	if categoryID != nil {
		catUUID, err := uuid.FromString(*categoryID)
		if err != nil {
			return nil, err
		}
		query = query.Joins("JOIN category_products cp ON cp.product_id = products.id").
			Where("cp.category_id = ?", catUUID)
	}

	// Apply search filter if provided
	if search != nil && *search != "" {
		searchTerm := "%" + *search + "%"
		query = query.Where(
			"products.name ILIKE ? OR products.description ILIKE ? OR products.sku ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	// Convert to GraphQL type
	result := make([]*model.Product, len(products))
	for i, product := range products {
		result[i] = product.ToGraphQL()
	}

	return result, nil
}

func GetCategoryAveragePrice(categoryID string) (float64, error) {
	catUUID, err := uuid.FromString(categoryID)
	if err != nil {
		return 0, err
	}

	var avgPrice float64
	err = utils.DB.Model(&models.Product{}).
		Joins("JOIN category_products cp ON cp.product_id = products.id").
		Where("cp.category_id = ?", catUUID).
		Select("AVG(price)").
		Row().
		Scan(&avgPrice)

	if err != nil {
		return 0, err
	}

	return avgPrice, nil
}

// implement this UpdateProduct(id, input)
func UpdateProduct(id string, input model.ProductInput) (*model.Product, error) {
	productUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}
	var product models.Product
	if err := utils.DB.First(&product, "id = ?", productUUID).Error; err != nil {
		return nil, err
	}
	product.Name = input.Name
	if input.Description != nil {
		product.Description = *input.Description
	}
	product.Price = input.Price
	product.Stock = int(input.Stock)
	if err := utils.DB.Save(&product).Error; err != nil {
		return nil, err
	}
	return product.ToGraphQL(), nil
}

// implement this DeleteProduct(id)
func DeleteProduct(id string) (bool, error) {
	productUUID, err := uuid.FromString(id)
	if err != nil {
		return false, err
	}
	var product models.Product
	if err := utils.DB.First(&product, "id = ?", productUUID).Error; err != nil {
		return false, err
	}
	if err := utils.DB.Delete(&product).Error; err != nil {
		return false, err
	}
	return true, nil

}

// implement this GetProduct(id)
func GetProductByID(id string) (*model.Product, error) {
	productUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}
	var product models.Product
	if err := utils.DB.Preload("Categories").First(&product, "id = ?", productUUID).Error; err != nil {
		return nil, err
	}
	return product.ToGraphQL(), nil
}
