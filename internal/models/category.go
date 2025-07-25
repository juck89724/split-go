package models

// Category 交易分類
type Category struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"not null"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}
