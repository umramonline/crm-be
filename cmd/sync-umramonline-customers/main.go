package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	customerpersistence "github.com/umran/new.crm/backend/internal/customer/infrastructure/persistence"
	"github.com/umran/new.crm/backend/internal/infrastructure/config"
	dbpersistence "github.com/umran/new.crm/backend/internal/infrastructure/persistence"
	"gorm.io/gorm"
)

const defaultBatchSize = 500

type sourceCustomer struct {
	ID               uint64
	BranchID         *int32
	Unvan            *string
	Ad               *string
	Soyad            *string
	YetkiliAdi       *string `gorm:"column:yetkili_adi"`
	Cep              *string
	Telefon          *string
	Fax              *string
	Eposta           *string
	Web              *string
	Mahalle          *string
	Cadde            *string
	Sokak            *string
	Semt             *string
	IlKodu           *string `gorm:"column:il_kodu"`
	IlceKodu         *string `gorm:"column:ilce_kodu"`
	Ulke             *string
	DogumTarihi      *string `gorm:"column:dogum_tarihi"`
	VadeGunu         *string `gorm:"column:vade_gunu"`
	VergiDairesi     *string `gorm:"column:vergi_dairesi"`
	VergiDairesiKodu *string `gorm:"column:vergi_dairesi_kodu"`
	VergiNo          *string `gorm:"column:vergi_no"`
	TCNo             *string `gorm:"column:tc_no"`
	Type             *string
	CreatedAt        *string
	UpdatedAt        *string
	Mersis           *string
	PasaportNo       *string `gorm:"column:pasaport_no"`
	PasaportBelge    *string `gorm:"column:pasaport_belge"`
	EsbisNo          *string `gorm:"column:esbis_no"`
	YetkiBelgeNo     *string `gorm:"column:yetki_belge_no"`
	KapiNo           *string `gorm:"column:kapi_no"`
}

func (sourceCustomer) TableName() string {
	return "customers"
}

func main() {
	_ = godotenv.Load("../umramonline/.env.local", "../umramonline/.env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		log.Fatal("DATABASE_DSN is required for backend database")
	}

	backendDB, err := dbpersistence.OpenMySQL(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("open backend database: %v", err)
	}

	umramonlineDB, err := dbpersistence.OpenMySQL("root:root@tcp(127.0.0.1:33007)/umramdb")
	if err != nil {
		log.Fatalf("open umramonline database: %v", err)
	}

	batchSize := envInt("SYNC_BATCH_SIZE", defaultBatchSize)
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	stats, err := syncCustomers(umramonlineDB, backendDB, batchSize)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"customer sync completed scanned=%d inserted=%d updated=%d",
		stats.Scanned,
		stats.Inserted,
		stats.Updated,
	)
}

type syncStats struct {
	Scanned  int
	Inserted int
	Updated  int
}

func syncCustomers(sourceDB *gorm.DB, targetDB *gorm.DB, batchSize int) (syncStats, error) {
	stats := syncStats{}
	var lastID uint64

	for {
		customers := []sourceCustomer{}
		if err := sourceDB.
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&customers).Error; err != nil {
			return stats, fmt.Errorf("read umramonline customers: %w", err)
		}

		if len(customers) == 0 {
			return stats, nil
		}

		if err := targetDB.Transaction(func(tx *gorm.DB) error {
			for _, customer := range customers {
				inserted, err := upsertCustomer(tx, customer)
				if err != nil {
					return err
				}

				stats.Scanned++
				if inserted {
					stats.Inserted++
				} else {
					stats.Updated++
				}
			}

			return nil
		}); err != nil {
			return stats, err
		}

		lastID = customers[len(customers)-1].ID
	}
}

func upsertCustomer(db *gorm.DB, source sourceCustomer) (bool, error) {
	if source.ID == 0 {
		return false, errors.New("umramonline customer id is empty")
	}

	target := targetCustomer(source)

	var existing customerpersistence.CustomerModel
	err := db.Where("uo_id = ?", source.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&target).Error; err != nil {
				return false, fmt.Errorf("insert customer uo_id=%d: %w", source.ID, err)
			}

			return true, nil
		}

		return false, fmt.Errorf("find customer uo_id=%d: %w", source.ID, err)
	}

	if err := db.Model(&existing).Updates(targetCustomerUpdates(source)).Error; err != nil {
		return false, fmt.Errorf("update customer uo_id=%d: %w", source.ID, err)
	}

	return false, nil
}

func datePointer(value *string) *time.Time {
	if value == nil {
		return nil
	}

	date, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil
	}

	return &date
}

func targetCustomer(source sourceCustomer) customerpersistence.CustomerModel {
	return customerpersistence.CustomerModel{
		UOId:             source.ID,
		BranchID:         source.BranchID,
		Unvan:            trimStringPointer(source.Unvan),
		Ad:               trimStringPointer(source.Ad),
		Soyad:            trimStringPointer(source.Soyad),
		YetkiliAdi:       trimStringPointer(source.YetkiliAdi),
		Cep:              trimStringPointer(source.Cep),
		Telefon:          trimStringPointer(source.Telefon),
		Fax:              trimStringPointer(source.Fax),
		Eposta:           trimStringPointer(source.Eposta),
		Web:              trimStringPointer(source.Web),
		Mahalle:          trimStringPointer(source.Mahalle),
		Cadde:            trimStringPointer(source.Cadde),
		Sokak:            trimStringPointer(source.Sokak),
		Semt:             trimStringPointer(source.Semt),
		IlKodu:           trimStringPointer(source.IlKodu),
		IlceKodu:         trimStringPointer(source.IlceKodu),
		Ulke:             trimStringPointer(source.Ulke),
		DogumTarihi:      datePointer(source.DogumTarihi),
		VadeGunu:         datePointer(source.VadeGunu),
		VergiDairesi:     trimStringPointer(source.VergiDairesi),
		VergiDairesiKodu: trimStringPointer(source.VergiDairesiKodu),
		VergiNo:          trimStringPointer(source.VergiNo),
		TCNo:             trimStringPointer(source.TCNo),
		Type:             trimStringPointer(source.Type),
		CreatedAt:        nonZeroTime(source.CreatedAt),
		UpdatedAt:        nonZeroTime(source.UpdatedAt),
		Mersis:           trimStringPointer(source.Mersis),
		PasaportNo:       trimStringPointer(source.PasaportNo),
		PasaportBelge:    trimStringPointer(source.PasaportBelge),
		EsbisNo:          trimStringPointer(source.EsbisNo),
		YetkiBelgeNo:     trimStringPointer(source.YetkiBelgeNo),
		KapiNo:           trimStringPointer(source.KapiNo),
	}
}

func targetCustomerUpdates(source sourceCustomer) map[string]any {
	return map[string]any{
		"branch_id":          source.BranchID,
		"unvan":              trimStringPointer(source.Unvan),
		"ad":                 trimStringPointer(source.Ad),
		"soyad":              trimStringPointer(source.Soyad),
		"yetkili_adi":        trimStringPointer(source.YetkiliAdi),
		"cep":                trimStringPointer(source.Cep),
		"telefon":            trimStringPointer(source.Telefon),
		"fax":                trimStringPointer(source.Fax),
		"eposta":             trimStringPointer(source.Eposta),
		"web":                trimStringPointer(source.Web),
		"mahalle":            trimStringPointer(source.Mahalle),
		"cadde":              trimStringPointer(source.Cadde),
		"sokak":              trimStringPointer(source.Sokak),
		"semt":               trimStringPointer(source.Semt),
		"il_kodu":            trimStringPointer(source.IlKodu),
		"ilce_kodu":          trimStringPointer(source.IlceKodu),
		"ulke":               trimStringPointer(source.Ulke),
		"dogum_tarihi":       source.DogumTarihi,
		"vade_gunu":          source.VadeGunu,
		"vergi_dairesi":      trimStringPointer(source.VergiDairesi),
		"vergi_dairesi_kodu": trimStringPointer(source.VergiDairesiKodu),
		"vergi_no":           trimStringPointer(source.VergiNo),
		"tc_no":              trimStringPointer(source.TCNo),
		"type":               trimStringPointer(source.Type),
		"updated_at":         nonZeroTime(source.UpdatedAt),
		"mersis":             trimStringPointer(source.Mersis),
		"pasaport_no":        trimStringPointer(source.PasaportNo),
		"pasaport_belge":     trimStringPointer(source.PasaportBelge),
		"esbis_no":           trimStringPointer(source.EsbisNo),
		"yetki_belge_no":     trimStringPointer(source.YetkiBelgeNo),
		"kapi_no":            trimStringPointer(source.KapiNo),
	}
}

func trimStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

func nonZeroTime(value *string) time.Time {
	if value == nil {
		return time.Time{}
	}

	date, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return time.Time{}
	}

	return date
}

func envInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsedValue
}
