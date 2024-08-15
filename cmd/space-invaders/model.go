package main

import (
	"fmt"
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
	Count    int64  `yaml:"count" json:"count"`
}

// Score represents a player's score.
type Score struct {
	BaseModel
	Name  string `yaml:"name" json:"name" gorm:"primaryKey"`
	Score int64  `yaml:"score" json:"score"`
}

// Size represents a raw byte size.
type Size int64

// String returns the size as a human-readable string.
func (s Size) String() string {
	const unit = 1024
	if s < unit {
		return fmt.Sprintf("%d B", s)
	}

	div, exp := int64(unit), 0
	for n := s / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(s)/float64(div), "KMGTPE"[exp])
}
