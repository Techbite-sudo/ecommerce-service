package models

// Customer represents a customer in the system
type Customer struct {
	Base
	Name        string  `gorm:"not null"`
	Email       string  `gorm:"not null;unique"`
	Password    string  `gorm:"not null"`
	PhoneNumber string  `gorm:"not null"`
	Orders      []Order `gorm:"foreignkey:CustomerID"`
}
