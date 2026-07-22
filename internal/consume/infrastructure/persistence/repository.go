package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	customerpersistence "github.com/umran/new.crm/backend/internal/customer/infrastructure/persistence"
	"github.com/umran/new.crm/backend/internal/consume/application"
	"github.com/umran/new.crm/backend/internal/consume/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ConsumeCustomerCreated(ctx context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error) {
	if r == nil || r.db == nil {
		return domain.ConsumeResult{}, gorm.ErrInvalidDB
	}

	if processed, err := r.isEventProcessed(ctx, event.EventID); err != nil {
		return domain.ConsumeResult{}, err
	} else if processed {
		return domain.ConsumeResult{
			EventID: event.EventID,
			Action:  "already_processed",
		}, nil
	}

	var result domain.ConsumeResult

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if processed, err := r.isEventProcessedTx(tx, event.EventID); err != nil {
			return err
		} else if processed {
			result = domain.ConsumeResult{
				EventID: event.EventID,
				Action:  "already_processed",
			}

			return nil
		}

		customerID, action, err := r.upsertCustomerFromEvent(tx, event)
		if err != nil {
			return err
		}

		if err := tx.Create(&ProcessedEventModel{EventUUID: event.EventID}).Error; err != nil {
			if isDuplicateKeyError(err) {
				result = domain.ConsumeResult{
					EventID: event.EventID,
					Action:  "already_processed",
				}

				return nil
			}

			return err
		}

		result = domain.ConsumeResult{
			EventID:    event.EventID,
			CustomerID: customerID,
			Action:     action,
		}

		return nil
	})
	if err != nil {
		return domain.ConsumeResult{}, err
	}

	return result, nil
}

func (r *Repository) upsertCustomerFromEvent(tx *gorm.DB, event domain.CustomerCreatedEvent) (uint64, string, error) {
	existing, found, err := findDuplicateCustomer(tx, event)
	if err != nil {
		return 0, "", err
	}

	model := customerModelFromEvent(event)

	if found {
		if err := tx.Model(&existing).Updates(customerUpdatesFromEvent(event)).Error; err != nil {
			return 0, "", err
		}

		if err := replaceCustomerTelephones(tx, existing.ID, event.Telephones); err != nil {
			return 0, "", err
		}

		return existing.ID, "updated", nil
	}

	if err := tx.Create(&model).Error; err != nil {
		return 0, "", err
	}

	if err := replaceCustomerTelephones(tx, model.ID, event.Telephones); err != nil {
		return 0, "", err
	}

	return model.ID, "created", nil
}

func findDuplicateCustomer(tx *gorm.DB, event domain.CustomerCreatedEvent) (customerpersistence.CustomerModel, bool, error) {
	identity := strings.TrimSpace(event.TCNo)
	if identity == "" {
		identity = strings.TrimSpace(event.VergiNo)
	}

	telefon := strings.TrimSpace(event.Telefon)
	cep := strings.TrimSpace(event.Cep)

	var customer customerpersistence.CustomerModel
	err := tx.
		Where("(tc_no = ? OR vergi_no = ?)", identity, identity).
		Where("(telefon = ? OR cep = ? OR telefon = ? OR cep = ?)", telefon, telefon, cep, cep).
		First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerpersistence.CustomerModel{}, false, nil
		}

		return customerpersistence.CustomerModel{}, false, err
	}

	return customer, true, nil
}

func customerModelFromEvent(event domain.CustomerCreatedEvent) customerpersistence.CustomerModel {
	createdAt := parseTimestamp(event.CreatedAt)
	updatedAt := parseTimestamp(event.UpdatedAt)

	return customerpersistence.CustomerModel{
		UOId:             event.UOId,
		BranchID:         branchIDPointer(event.BranchID),
		Unvan:            stringPointer(event.Unvan),
		Ad:               stringPointer(event.Ad),
		Soyad:            stringPointer(event.Soyad),
		YetkiliAdi:       stringPointer(event.YetkiliAdi),
		Cep:              stringPointer(event.Cep),
		Telefon:          stringPointer(event.Telefon),
		Fax:              stringPointer(event.Fax),
		Eposta:           stringPointer(event.Eposta),
		Web:              stringPointer(event.Web),
		Mahalle:          stringPointer(event.Mahalle),
		Cadde:            stringPointer(event.Cadde),
		Sokak:            stringPointer(event.Sokak),
		Semt:             stringPointer(event.Semt),
		KapiNo:           stringPointer(event.KapiNo),
		IlKodu:           stringPointer(event.IlKodu),
		IlceKodu:         stringPointer(event.IlceKodu),
		Ulke:             stringPointer(event.Ulke),
		DogumTarihi:      parseOptionalDate(event.DogumTarihi),
		VadeGunu:         parseOptionalDate(event.VadeGunu),
		VergiDairesi:     stringPointer(event.VergiDairesi),
		VergiDairesiKodu: stringPointer(event.VergiDairesiKodu),
		VergiNo:          stringPointer(event.VergiNo),
		TCNo:             stringPointer(event.TCNo),
		Type:             stringPointer(defaultCustomerType(event.Type)),
		Mersis:           stringPointer(event.Mersis),
		PasaportNo:       stringPointer(event.PasaportNo),
		PasaportBelge:    stringPointer(event.PasaportBelge),
		EsbisNo:          stringPointer(event.EsbisNo),
		YetkiBelgeNo:     stringPointer(event.YetkiBelgeNo),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

func customerUpdatesFromEvent(event domain.CustomerCreatedEvent) map[string]interface{} {
	updates := map[string]interface{}{
		"uo_id":              event.UOId,
		"branch_id":          branchIDPointer(event.BranchID),
		"unvan":              stringPointer(event.Unvan),
		"ad":                 stringPointer(event.Ad),
		"soyad":              stringPointer(event.Soyad),
		"yetkili_adi":        stringPointer(event.YetkiliAdi),
		"cep":                stringPointer(event.Cep),
		"telefon":            stringPointer(event.Telefon),
		"fax":                stringPointer(event.Fax),
		"eposta":             stringPointer(event.Eposta),
		"web":                stringPointer(event.Web),
		"mahalle":            stringPointer(event.Mahalle),
		"cadde":              stringPointer(event.Cadde),
		"sokak":              stringPointer(event.Sokak),
		"semt":               stringPointer(event.Semt),
		"kapi_no":            stringPointer(event.KapiNo),
		"il_kodu":            stringPointer(event.IlKodu),
		"ilce_kodu":          stringPointer(event.IlceKodu),
		"ulke":               stringPointer(event.Ulke),
		"dogum_tarihi":       parseOptionalDate(event.DogumTarihi),
		"vade_gunu":          parseOptionalDate(event.VadeGunu),
		"vergi_dairesi":      stringPointer(event.VergiDairesi),
		"vergi_dairesi_kodu": stringPointer(event.VergiDairesiKodu),
		"vergi_no":           stringPointer(event.VergiNo),
		"tc_no":              stringPointer(event.TCNo),
		"type":               stringPointer(defaultCustomerType(event.Type)),
		"mersis":             stringPointer(event.Mersis),
		"pasaport_no":        stringPointer(event.PasaportNo),
		"pasaport_belge":     stringPointer(event.PasaportBelge),
		"esbis_no":           stringPointer(event.EsbisNo),
		"yetki_belge_no":     stringPointer(event.YetkiBelgeNo),
		"updated_at":         parseTimestamp(event.UpdatedAt),
	}

	return updates
}

func replaceCustomerTelephones(tx *gorm.DB, customerID uint64, telephones []domain.Telephone) error {
	if err := tx.Where("customer_id = ?", customerID).Delete(&customerpersistence.CustomerTelephoneModel{}).Error; err != nil {
		return err
	}

	for _, telephone := range telephones {
		phoneNumber := strings.TrimSpace(telephone.PhoneNumber)
		title := strings.TrimSpace(telephone.Title)
		if phoneNumber == "" && title == "" {
			continue
		}

		model := customerpersistence.CustomerTelephoneModel{
			CustomerID:  customerID,
			PhoneNumber: phoneNumber,
			Title:       stringPointer(title),
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) isEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&ProcessedEventModel{}).
		Where("event_uuid = ?", eventID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) isEventProcessedTx(tx *gorm.DB, eventID string) (bool, error) {
	var count int64
	if err := tx.Model(&ProcessedEventModel{}).Where("event_uuid = ?", eventID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func branchIDPointer(branchID int32) *int32 {
	if branchID == 0 {
		return nil
	}

	value := branchID

	return &value
}

func defaultCustomerType(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "bireysel"
	}

	return trimmedValue
}

func stringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

func parseOptionalDate(value *string) *time.Time {
	if value == nil {
		return nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" || strings.EqualFold(trimmedValue, "null") {
		return nil
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
	}

	for _, layout := range layouts {
		parsedValue, err := time.Parse(layout, trimmedValue)
		if err == nil {
			return &parsedValue
		}
	}

	return nil
}

func parseTimestamp(value string) time.Time {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return time.Now()
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsedValue, err := time.Parse(layout, trimmedValue)
		if err == nil {
			return parsedValue
		}
	}

	return time.Now()
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())

	return strings.Contains(message, "duplicate") || strings.Contains(message, "1062")
}

var _ application.Repository = (*Repository)(nil)
