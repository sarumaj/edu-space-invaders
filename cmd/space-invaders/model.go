package main

import (
	"time"
)

// Base is the base model for the database models.
type Base struct {
	CreatedAt time.Time `yaml:"-" json:"-"`
	UpdatedAt time.Time `yaml:"-" json:"-"`
}

// Score represents a player's score.
type Score struct {
	Base
	Name  string `yaml:"name" json:"name" gorm:"primaryKey"`
	Score int    `yaml:"score" json:"score"`
}
