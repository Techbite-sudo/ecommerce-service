package middleware

import (
	// "context"
	"ecommerce-service/graph/model"
	"fmt"
	"regexp"
	"strings"
)

var (
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	skuRegex   = regexp.MustCompile(`^[A-Za-z0-9-_]+$`)
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate product input
func ValidateProductInput(input model.ProductInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}

	if input.Price <= 0 {
		return &ValidationError{Field: "price", Message: "must be greater than 0"}
	}

	if !skuRegex.MatchString(input.Sku) {
		return &ValidationError{Field: "sku", Message: "invalid format"}
	}

	if input.Stock < 0 {
		return &ValidationError{Field: "stock", Message: "cannot be negative"}
	}

	return nil
}

// Validate category input
func ValidateCategoryInput(input model.CategoryInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}

	return nil
}

// Validate order input
func ValidateOrderInput(input model.OrderInput) error {
	if len(input.Items) == 0 {
		return &ValidationError{Field: "items", Message: "order must contain at least one item"}
	}

	for i, item := range input.Items {
		if item.Quantity <= 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("items[%d].quantity", i),
				Message: "must be greater than 0",
			}
		}
	}

	return nil
}

// Validate profile update input
func ValidateUpdateProfileInput(input model.UpdateProfileInput) error {
	if input.PhoneNumber != nil {
		if !phoneRegex.MatchString(*input.PhoneNumber) {
			return &ValidationError{Field: "phoneNumber", Message: "invalid phone number format"}
		}
	}

	if input.Country != nil {
		if strings.TrimSpace(*input.Country) == "" {
			return &ValidationError{Field: "country", Message: "cannot be empty"}
		}
	}

	return nil
}

// // Add validation directives to GraphQL schema
// func ValidateDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
//     // Implementation depends on your GraphQL framework
//     return next(ctx)
// }
