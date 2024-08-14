package main

import (
	"time"
)

// BaseModel is the base model for the database models.
type BaseModel struct {
	CreatedAt *time.Time `yaml:"created_at,omitempty" json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt *time.Time `yaml:"updated_at,omitempty" json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

// Metric represents a metrics entry.
type Metric struct {
	BaseModel
	Endpoint string `yaml:"endpoint" json:"endpoint" gorm:"primaryKey"`
	Method   string `yaml:"method" json:"method" gorm:"primaryKey"`
	Count    int    `yaml:"count" json:"count"`
}

// Score represents a player's score.
type Score struct {
	BaseModel
	Name  string `yaml:"name" json:"name" gorm:"primaryKey"`
	Score int    `yaml:"score" json:"score"`
}
