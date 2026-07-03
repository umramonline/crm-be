package persistence

import (
	"time"

	"gorm.io/gorm"
)

type CustomerModel struct {
	ID                     uint64     `gorm:"primaryKey;autoIncrement"`
	UOId                   uint64     `gorm:"column:uo_id;type:bigint"`
	BranchID               *int32     `gorm:"column:branch_id;type:int"`
	Unvan                  *string    `gorm:"size:255"`
	Ad                     *string    `gorm:"size:255"`
	Soyad                  *string    `gorm:"size:255"`
	YetkiliAdi             *string    `gorm:"column:yetkili_adi;size:255"`
	Cep                    *string    `gorm:"size:255"`
	Telefon                *string    `gorm:"size:255"`
	Fax                    *string    `gorm:"size:255"`
	Eposta                 *string    `gorm:"size:255"`
	Web                    *string    `gorm:"size:255"`
	Mahalle                *string    `gorm:"size:255"`
	Cadde                  *string    `gorm:"size:255"`
	Sokak                  *string    `gorm:"size:255"`
	Semt                   *string    `gorm:"size:255"`
	IlKodu                 *string    `gorm:"column:il_kodu;size:255"`
	IlceKodu               *string    `gorm:"column:ilce_kodu;size:255"`
	Ulke                   *string    `gorm:"size:255"`
	AddressDetail          *string    `gorm:"column:address_detail;size:255"`
	DogumTarihi            *time.Time `gorm:"column:dogum_tarihi;type:date"`
	VadeGunu               *time.Time `gorm:"column:vade_gunu;type:date"`
	VergiDairesi           *string    `gorm:"column:vergi_dairesi;size:255"`
	VehicleStockCount      *int32     `gorm:"column:vehicle_stock_count;type:int"`
	VergiDairesiKodu       *string    `gorm:"column:vergi_dairesi_kodu;size:255"`
	VergiNo                *string    `gorm:"column:vergi_no;size:255"`
	TCNo                   *string    `gorm:"column:tc_no;size:255"`
	Type                   *string    `gorm:"size:255;default:bireysel"`
	CreatedAt              time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt              time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt              gorm.DeletedAt
	Mersis                 *string `gorm:"size:20"`
	PasaportNo             *string `gorm:"column:pasaport_no;size:50"`
	PasaportBelge          *string `gorm:"column:pasaport_belge;size:255"`
	EsbisNo                *string `gorm:"column:esbis_no;size:255"`
	YetkiBelgeNo           *string `gorm:"column:yetki_belge_no;size:255"`
	KapiNo                 *string `gorm:"column:kapi_no;size:255"`
	Website                *string `gorm:"column:website;type:varchar(255)"`
	GoogleMapLink          *string `gorm:"column:google_map_link;type:varchar(255)"`
	ClassifiedsWebsiteLink *string `gorm:"column:classifieds_website_link;type:varchar(255)"`
	CorporateSector        *string `gorm:"column:corporate_sector;type:varchar(255)"`
}

type CustomerTelephoneModel struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	CustomerID  uint64    `gorm:"column:customer_id;type:bigint;not null;index"`
	PhoneNumber string    `gorm:"column:phone_number;size:255;not null"`
	Title       *string   `gorm:"size:255"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt
}

type TaskModel struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement"`
	UUID           string     `gorm:"column:uuid;type:char(36);not null;uniqueIndex"`
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt      *time.Time `gorm:"type:timestamp;index"`
	Title          string     `gorm:"size:255;not null"`
	Description    *string    `gorm:"size:255"`
	AssignedUserID uint64     `gorm:"column:assigned_user_id;type:bigint unsigned;not null;index"`
	BranchID       uint64     `gorm:"column:branch_id;type:bigint unsigned;not null;index"`
	VisitDate      *time.Time `gorm:"column:visit_date;type:date"`
	DueDate        *time.Time `gorm:"column:due_date;type:date"`
	Status         string     `gorm:"type:enum('pending','in_progress','cancelled');not null;default:pending"`
	Priority       string     `gorm:"type:enum('high','medium','low');not null;default:medium"`
}

type TaskCustomerModel struct {
	TaskID     uint64        `gorm:"column:task_id;type:bigint unsigned;not null;primaryKey"`
	CustomerID uint64        `gorm:"column:customer_id;type:bigint unsigned;not null;primaryKey"`
	CreatedAt  time.Time     `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Task       TaskModel     `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	Customer   CustomerModel `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
}

func (CustomerTelephoneModel) TableName() string {
	return "customer_telephones"
}

func (CustomerModel) TableName() string {
	return "customers"
}

func (TaskModel) TableName() string {
	return "tasks"
}

func (TaskCustomerModel) TableName() string {
	return "tasks_customers"
}

func AutoMigrate(db *gorm.DB) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").
		AutoMigrate(&CustomerModel{}, &CustomerTelephoneModel{}, &TaskModel{}, &TaskCustomerModel{})
}
