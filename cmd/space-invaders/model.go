package main

import (
	"fmt"
	"math"
	"time"

	gorm "gorm.io/gorm"
	clause "gorm.io/gorm/clause"
)

const maximumSize = 1 << 30         // 1 GiB
const sizeThreshold = 1_000_000_000 // 1 GB, approximately 93% of the maximum size

const databaseSizeQuery = "SELECT pg_database_size(current_database())"
const tableSizeQuery = "SELECT pg_total_relation_size(?)"

// helper is a helper for the database.
type helper struct{ *gorm.DB }

// ClearMetrics clears the metrics.
// It keeps the most recently updated metrics specified by keepTopMostRecent.
func (database helper) ClearMetrics(keepTopMostRecent int) error {
	subQuery := database.
		Model(&Metric{}).
		Select("endpoint", "method").
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: "CASE WHEN updated_at > created_at THEN updated_at ELSE created_at END", Raw: true},
			Desc:   true,
		}).
		Offset(keepTopMostRecent)
	return database.
		Where("(endpoint, method) IN (?)", subQuery).
		Delete(&Score{}).
		Error
}

// ClearMetrics clears the metrics.
// It keeps the metrics with the highest count specified by keepTopMetrics.
func (database helper) ClearScores(keepTopScores int) error {
	subQuery := database.
		Model(&Score{}).
		Select("name").
		Order(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: "score"}, Desc: true},
				{Column: clause.Column{Name: "CASE WHEN updated_at > created_at THEN updated_at ELSE created_at END", Raw: true}},
			},
		}).
		Offset(keepTopScores)
	return database.
		Where("name IN (?)", subQuery).
		Delete(&Score{}).
		Error
}

// GetDatabaseSize returns the database size.
func (database helper) GetDatabaseSize() (Size, error) {
	var size int64
	if err := database.Raw(databaseSizeQuery).Scan(&size).Error; err != nil {
		return 0, err
	}

	return Size(size), nil
}

// GetTableSizes returns the table sizes.
// It returns the table sizes as a map of table names to sizes.
func (database helper) GetTableSizes() (map[string]Size, error) {
	tables, err := database.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	sizes := make(map[string]Size, len(tables))
	for _, table := range tables {
		var size int64
		if err := database.Raw(tableSizeQuery, table).Scan(&size).Error; err != nil {
			return nil, err
		}
		sizes[table] = Size(size)
	}

	return sizes, nil
}

// GetMetrics returns the metrics.
// It returns the metrics sorted by count in descending order.
func (database helper) GetScores() ([]Score, error) {
	scores := make([]Score, 0)
	if err := database.
		Order(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: "score"}, Desc: true},
				{Column: clause.Column{Name: "CASE WHEN updated_at > created_at THEN updated_at ELSE created_at END", Raw: true}},
			},
		}).
		Find(&scores).
		Error; err != nil {

		return nil, err
	}

	return scores, nil
}

// SaveMetric saves the metric.
// It increments the count if the metric already exists.
// It updates the updated_at field.
func (database helper) SaveMetric(metric Metric) error {
	return database.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "endpoint"}, {Name: "method"}},
			DoUpdates: clause.Assignments(map[string]any{
				"count":      gorm.Expr("CASE WHEN metrics.count < ? THEN metrics.count + ? ELSE metrics.count END", math.MaxInt64, 1), // Increment the count.
				"updated_at": gorm.Expr("?", time.Now()),
			}),
			Where: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("EXCLUDED.endpoint = metrics.endpoint"),
				gorm.Expr("EXCLUDED.method = metrics.method"),
			}},
		}).
		Create([]Metric{metric}).
		Error
}

// SaveScores saves the scores.
// It updates the score if the new score is higher.
// It updates the updated_at field.
func (database helper) SaveScores(scores []Score) error {
	if len(scores) == 0 {
		return database.Where("1 = 1").Delete(&Score{}).Error
	}

	return database.
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"score":      gorm.Expr("CASE WHEN EXCLUDED.score < ? THEN EXCLUDED.score ELSE scores.score END", math.MaxInt64),
				"updated_at": gorm.Expr("?", time.Now()),
			}),
			Where: clause.Where{Exprs: []clause.Expression{gorm.Expr("EXCLUDED.score > scores.score")}},
		}).
		Create(scores).
		Error
}

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

// Helper returns a helper for the database.
func Helper(database *gorm.DB) helper {
	return helper{database}
}
