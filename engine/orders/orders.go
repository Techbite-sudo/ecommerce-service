package orders

import (
	"ecommerce-service/engine/notifications"
	"ecommerce-service/graph/model"
	"ecommerce-service/models"
	"ecommerce-service/utils"
	"errors"

	uuid "github.com/satori/go.uuid"
)

func CreateOrder(input model.OrderInput, userID string) (*model.Order, error) {
	// Start transaction
	tx := utils.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return nil, err
	}

	// Create order
	order := models.Order{
		CustomerID: userUUID,
		Status:     models.OrderStatusPending,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Process order items
	var total float64
	for _, itemInput := range input.Items {
		productUUID, err := uuid.FromString(itemInput.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Get product
		var product models.Product
		if err := tx.First(&product, "id = ?", productUUID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Check stock
		if int(product.Stock) < int(itemInput.Quantity) {
			tx.Rollback()
			return nil, errors.New("insufficient stock for product: " + product.Name)
		}

		// Create order item
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: productUUID,
			Quantity:  int(itemInput.Quantity),
			UnitPrice: product.Price,
			SubTotal:  product.Price * float64(itemInput.Quantity),
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update product stock
		product.Stock -= int(itemInput.Quantity)
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		total += orderItem.SubTotal
	}

	// Update order total
	order.Total = total
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load full order with relations
	if err := utils.DB.Preload("Customer").Preload("Items.Product").First(&order, order.ID).Error; err != nil {
		return nil, err
	}

	// Send notifications
	go func() {
		notifications.SendOrderConfirmationSMS(&order)
		notifications.SendOrderNotificationEmail(&order)
	}()

	return order.ToGraphQL(), nil
}
func GetOrder(id string, userID string) (*model.Order, error) {
	orderUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	var order models.Order
	if err := utils.DB.Preload("Customer").Preload("Items.Product").First(&order, "id = ?", orderUUID).Error; err != nil {
		return nil, err
	}

	// Verify user owns this order
	if order.CustomerID.String() != userID && order.Customer.Role != models.RoleAdmin {
		return nil, errors.New("unauthorized access to order")
	}

	return order.ToGraphQL(), nil
}

func GetUserOrders(userID string) ([]*model.Order, error) {
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := utils.DB.Preload("Customer").Preload("Items.Product").
		Where("customer_id = ?", userUUID).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	result := make([]*model.Order, len(orders))
	for i, order := range orders {
		result[i] = order.ToGraphQL()
	}

	return result, nil
}
