package persistence

import "time"

type TaskModel struct {
	ID                    uint64     `gorm:"primaryKey;autoIncrement"`
	UUID                  string     `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	CreatedAt             time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt             *time.Time `gorm:"type:timestamp;index"`
	Title                 string     `gorm:"size:255;not null"`
	Description           *string    `gorm:"size:255"`
	CreatedByUserID       uint64     `gorm:"column:created_by_user_id;type:bigint unsigned;not null"`
	CreatedByUserFullName string     `gorm:"column:created_by_user_full_name;type:varchar(255);not null"`
	AssignedUserID        uint64     `gorm:"column:assigned_user_id;type:bigint unsigned;not null;index"`
	AssignedUserFullName  string     `gorm:"column:assigned_user_full_name;type:varchar(255);not null"`
	BranchID              uint64     `gorm:"column:branch_id;type:bigint unsigned;not null;index"`
	VisitDate             *time.Time `gorm:"column:visit_date;type:date"`
	DueDate               *time.Time `gorm:"column:due_date;type:date"`
	Status                string     `gorm:"type:enum('pending','in_progress','cancelled');not null;default:pending"`
	Priority              string     `gorm:"type:enum('high','medium','low');not null;default:medium"`
}

type TaskCustomerModel struct {
	TaskID     uint64        `gorm:"column:task_id;type:bigint unsigned;not null;primaryKey"`
	CustomerID uint64        `gorm:"column:customer_id;type:bigint unsigned;not null;primaryKey"`
	CreatedAt  time.Time     `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Task       TaskModel     `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	Customer   CustomerModel `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
}

type CustomerModel struct {
	ID       uint64  `gorm:"primaryKey;autoIncrement"`
	BranchID *uint64 `gorm:"column:branch_id;type:int"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (TaskCustomerModel) TableName() string {
	return "tasks_customers"
}

func (CustomerModel) TableName() string {
	return "customers"
}
