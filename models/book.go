package models

import (
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model         // adds fields, ID, createdAt, UpdatedAt, DeletedAt
	Title      string  `json:"title"`
	Author     string  `json:"author"`
	Price      float64 `json:"price"`
}
