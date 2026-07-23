package persistence

import (
	"time"

	"gorm.io/gorm"
)

type ProcessedEventModel struct {
	EventUUID   string    `gorm:"column:event_uuid;type:char(36);primaryKey"`
	UOId        uint64    `gorm:"column:uo_id;type:bigint;not null;index:idx_uo_event_occurred,priority:1"`
	EventType   string    `gorm:"column:event_type;type:varchar(100);not null;index:idx_uo_event_occurred,priority:2"`
	OccurredAt  time.Time `gorm:"column:occurred_at;type:timestamp;not null;index:idx_uo_event_occurred,priority:3"`
	ProcessedAt time.Time `gorm:"column:processed_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

func (ProcessedEventModel) TableName() string {
	return "processed_events"
}

func AutoMigrate(db *gorm.DB) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").
		AutoMigrate(&ProcessedEventModel{})
}
