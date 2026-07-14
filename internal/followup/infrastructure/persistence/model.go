package persistence

import "time"

type FollowUpModel struct {
	ID                     uint64            `gorm:"primaryKey;autoIncrement"`
	UUID                   string            `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	TasksCustomerID        uint64            `gorm:"column:tasks_customer_id;type:bigint unsigned;not null;index"`
	VisitType              string            `gorm:"column:visit_type;type:enum('Yerinde Ziyaret');not null"`
	VisitDate              time.Time         `gorm:"column:visit_date;type:timestamp;not null"`
	NextVisitDate          *time.Time        `gorm:"column:next_visit_date;type:timestamp"`
	AgreementReached       bool              `gorm:"column:agreement_reached;type:tinyint(1);not null;default:0"`
	AgreementFailureReason *string           `gorm:"column:agreement_failure_reason;type:enum('Fiyat yüksek','Mesafe Uzak','Bayi ile yaşanan sorunlar','Ekpertize ihtiyaç duymuyor','Kendisi yapıyor','Başka ekspertize yaptırıyor','Değerlendirme')"`
	Note                   *string           `gorm:"column:note;type:varchar(150)"`
	AssignedUserID         uint64            `gorm:"column:assigned_user_id;type:bigint unsigned;not null;index"`
	AssignedUserFullName   string            `gorm:"column:assigned_user_full_name;type:varchar(255);not null"`
	CreatedAt              *time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt              *time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt              *time.Time        `gorm:"type:timestamp;index"`
	TasksCustomer          TaskCustomerModel `gorm:"foreignKey:TasksCustomerID;constraint:OnDelete:RESTRICT"`
}

type FollowUpImageModel struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"`
	UUID            string        `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	TasksFollowUpID uint64        `gorm:"column:tasks_follow_up_id;type:bigint unsigned;not null;index"`
	Path            string        `gorm:"column:path;type:varchar(500);not null"`
	URL             string        `gorm:"column:url;type:varchar(500);not null"`
	FollowUp        FollowUpModel `gorm:"foreignKey:TasksFollowUpID;constraint:OnDelete:CASCADE"`
}

type MeetPersonModel struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"`
	UUID            string        `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	TasksFollowUpID uint64        `gorm:"column:tasks_follow_up_id;type:bigint unsigned;not null;index"`
	Title           string        `gorm:"column:title;type:enum('Genel Müdür','Satış Müdürü','Operasyon Müdürü','Pazarlama Müdürü','İşletme Müdürü','Bölge Müdürü','Şube Müdürü','Yönetici','Sahibi','Ortağı');not null"`
	Name            string        `gorm:"column:name;type:varchar(50);not null"`
	Surname         string        `gorm:"column:surname;type:varchar(50);not null"`
	Phone           string        `gorm:"column:phone;type:varchar(20);not null"`
	Email           *string       `gorm:"column:email;type:varchar(100)"`
	CreatedAt       *time.Time    `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time    `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt       *time.Time    `gorm:"type:timestamp;index"`
	FollowUp        FollowUpModel `gorm:"foreignKey:TasksFollowUpID;constraint:OnDelete:CASCADE"`
}

type TaskCustomerModel struct {
	ID         uint64  `gorm:"primaryKey;autoIncrement"`
	UUID       string  `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	TaskID     *uint64 `gorm:"column:task_id;type:bigint unsigned"`
	CustomerID uint64  `gorm:"column:customer_id;type:bigint unsigned;not null"`
	Status     string  `gorm:"column:status;type:enum('pending','in_progress','cancelled','completed');not null;default:pending"`
}

func (FollowUpModel) TableName() string {
	return "tasks_follow_ups"
}

func (FollowUpImageModel) TableName() string {
	return "tasks_follow_up_images"
}

func (MeetPersonModel) TableName() string {
	return "follows_meet_people"
}

func (TaskCustomerModel) TableName() string {
	return "tasks_customers"
}
