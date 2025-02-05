package models

import (
	"ecommerce-service/graph/model"
	"encoding/json"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Category struct {
	Base
	Name     string     `gorm:"not null"`
	ParentID *uuid.UUID `gorm:"type:uuid"`
	Parent   *Category  `gorm:"foreignkey:ParentID"`
	Children []Category `gorm:"foreignkey:ParentID"`
	Products []Product  `gorm:"many2many:category_products;"`
	Level    int        `gorm:"not null;default:0"` // For tracking hierarchy depth
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ParentID != nil {
		var parent Category
		if err := tx.First(&parent, "id = ?", c.ParentID).Error; err != nil {
			return err
		}
		c.Level = parent.Level + 1
	}
	return nil
}

func (c Category) MarshalJSON() ([]byte, error) {
	type CategoryAlias Category
	return json.Marshal(&struct {
		CategoryAlias
		ID string `json:"id"`
	}{
		CategoryAlias: CategoryAlias(c),
		ID:            c.ID.String(),
	})
}

// ToGraphQL
func (c Category) ToGraphQL() *model.Category {
	children := make([]*model.Category, len(c.Children))
	for i, child := range c.Children {
		children[i] = child.ToGraphQL()
	}
	return &model.Category{
		ID:       c.ID.String(),
		Name:     c.Name,
		Children: children,
	}
}
