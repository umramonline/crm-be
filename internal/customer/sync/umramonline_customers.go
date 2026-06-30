package customersync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	customerpersistence "github.com/umran/new.crm/backend/internal/customer/infrastructure/persistence"
	"gorm.io/gorm"
)

const DefaultBatchSize = 500

type SourceCustomer struct {
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
	DogumTarihi      *time.Time `gorm:"column:dogum_tarihi"`
	VadeGunu         *time.Time `gorm:"column:vade_gunu"`
	VergiDairesi     *string    `gorm:"column:vergi_dairesi"`
	VergiDairesiKodu *string    `gorm:"column:vergi_dairesi_kodu"`
	VergiNo          *string    `gorm:"column:vergi_no"`
	TCNo             *string    `gorm:"column:tc_no"`
	Type             *string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	Mersis           *string
	PasaportNo       *string `gorm:"column:pasaport_no"`
	PasaportBelge    *string `gorm:"column:pasaport_belge"`
	EsbisNo          *string `gorm:"column:esbis_no"`
	YetkiBelgeNo     *string `gorm:"column:yetki_belge_no"`
	KapiNo           *string `gorm:"column:kapi_no"`
}

func (SourceCustomer) TableName() string {
	return "customers"
}

type Stats struct {
	Scanned  int
	Inserted int
	Updated  int
}

type SchedulerConfig struct {
	SourceDB  *gorm.DB
	TargetDB  *gorm.DB
	BatchSize int
	DailyAt   string
	CronExpr  string
	Logger    *log.Logger
}

func YesterdayWindow(now time.Time) (time.Time, time.Time) {
	location := now.Location()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	startOfYesterday := startOfToday.AddDate(0, 0, -1)

	return startOfYesterday, startOfToday
}

func syncChangedCustomers(ctx context.Context, sourceDB *gorm.DB, targetDB *gorm.DB, batchSize int, startAt time.Time, endAt time.Time) (Stats, error) {
	stats := Stats{}
	if sourceDB == nil || targetDB == nil {
		return stats, errors.New("source and target database connections are required")
	}

	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	var lastID uint64
	for {
		if err := ctx.Err(); err != nil {
			return stats, err
		}

		customers := []SourceCustomer{}
		if err := sourceDB.WithContext(ctx).
			Where("id > ?", lastID).
			Where("(created_at >= ? AND created_at < ?) OR (updated_at >= ? AND updated_at < ?)", startAt, endAt, startAt, endAt).
			Order("id ASC").
			Limit(batchSize).
			Find(&customers).Error; err != nil {
			return stats, fmt.Errorf("read umramonline customers: %w", err)
		}

		if len(customers) == 0 {
			return stats, nil
		}

		if err := targetDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func StartDailyScheduler(ctx context.Context, config SchedulerConfig) {
	logger := config.Logger
	if logger == nil {
		logger = log.Default()
	}

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	schedule, err := cronSchedule(config.CronExpr, config.DailyAt)
	if err != nil {
		logger.Printf("customer sync scheduler disabled: %v", err)
		return
	}

	scheduler := cron.New(
		cron.WithLocation(time.Local),
		cron.WithChain(cron.SkipIfStillRunning(cron.PrintfLogger(logger))),
	)

	if _, err := scheduler.AddFunc(schedule, func() {
		startAt, endAt := YesterdayWindow(time.Now())
		logger.Printf("starting scheduled customer sync window=%s..%s", startAt.Format(time.RFC3339), endAt.Format(time.RFC3339))
		stats, err := syncChangedCustomers(ctx, config.SourceDB, config.TargetDB, batchSize, startAt, endAt)
		if err != nil {
			logger.Printf("scheduled customer sync failed: %v", err)
			return
		}

		logger.Printf("scheduled customer sync completed scanned=%d inserted=%d updated=%d", stats.Scanned, stats.Inserted, stats.Updated)
	}); err != nil {
		logger.Printf("customer sync scheduler disabled: %v", err)
		return
	}

	scheduler.Start()
	logger.Printf("customer sync cron scheduler started schedule=%q", schedule)

	go func() {
		<-ctx.Done()
		stopCtx := scheduler.Stop()
		<-stopCtx.Done()
	}()
}

func cronSchedule(cronExpr string, dailyAt string) (string, error) {
	normalizedCronExpr := strings.TrimSpace(cronExpr)
	if normalizedCronExpr != "" {
		return normalizedCronExpr, nil
	}

	normalizedValue := strings.TrimSpace(dailyAt)
	if normalizedValue == "" {
		normalizedValue = "03:00"
	}

	parsedTime, err := time.Parse("15:04", normalizedValue)
	if err != nil {
		return "", fmt.Errorf("CUSTOMER_SYNC_DAILY_AT must be HH:MM: %w", err)
	}

	return fmt.Sprintf("%d %d * * *", parsedTime.Minute(), parsedTime.Hour()), nil
}

func upsertCustomer(db *gorm.DB, source SourceCustomer) (bool, error) {
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

func targetCustomer(source SourceCustomer) customerpersistence.CustomerModel {
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
		DogumTarihi:      source.DogumTarihi,
		VadeGunu:         source.VadeGunu,
		VergiDairesi:     trimStringPointer(source.VergiDairesi),
		VergiDairesiKodu: trimStringPointer(source.VergiDairesiKodu),
		VergiNo:          trimStringPointer(source.VergiNo),
		TCNo:             trimStringPointer(source.TCNo),
		Type:             trimStringPointer(source.Type),
		CreatedAt:        timeValueOrNow(source.CreatedAt),
		UpdatedAt:        timeValueOrNow(source.UpdatedAt),
		Mersis:           trimStringPointer(source.Mersis),
		PasaportNo:       trimStringPointer(source.PasaportNo),
		PasaportBelge:    trimStringPointer(source.PasaportBelge),
		EsbisNo:          trimStringPointer(source.EsbisNo),
		YetkiBelgeNo:     trimStringPointer(source.YetkiBelgeNo),
		KapiNo:           trimStringPointer(source.KapiNo),
	}
}

func targetCustomerUpdates(source SourceCustomer) map[string]any {
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
		"updated_at":         timeValueOrNow(source.UpdatedAt),
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

func timeValueOrNow(value *time.Time) time.Time {
	if value == nil || value.IsZero() {
		return time.Now()
	}

	return *value
}
