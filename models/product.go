package models

// Product represents a product in the system
type Product struct {
	Base
	Name        string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Description string
	Categories  []Category `gorm:"many2many:product_categories;"`
}
