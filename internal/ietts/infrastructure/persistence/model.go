package persistence

import "time"

type IettsRecordModel struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	UUID              *string    `gorm:"column:uuid;type:char(36);uniqueIndex"`
	DocumentNumber    string     `gorm:"column:document_number;size:255;not null"`
	CompanyName       *string    `gorm:"column:company_name;size:255"`
	BusinessName      *string    `gorm:"column:business_name;size:255"`
	BusinessAddress   *string    `gorm:"column:business_address;size:255"`
	DocumentIssueDate *time.Time `gorm:"column:document_issue_date;type:date"`
	DocumentStatus    *string    `gorm:"column:document_status;size:255"`
	City              *string    `gorm:"column:city;size:255"`
	District          *string    `gorm:"column:district;size:255"`
	CustomerID        *uint64    `gorm:"column:customer_id;type:bigint unsigned"`
	CreatedAt         *time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt         *time.Time `gorm:"column:updated_at;type:timestamp"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;type:timestamp;index"`
}

func (IettsRecordModel) TableName() string {
	return "ietts_records"
}
