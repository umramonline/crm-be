package application

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umran/new.crm/backend/internal/followup/domain"
)

var ErrInvalidFollowUpCreateInput = errors.New("invalid follow up create input")

var ErrFollowUpCreateUnavailable = errors.New("follow up create unavailable")

var ErrFollowUpListUnavailable = errors.New("follow up list unavailable")

var ErrFollowUpDetailUnavailable = errors.New("follow up detail unavailable")

type ValidationErrors map[string]string

type Repository interface {
	FindTaskCustomerByUUID(ctx context.Context, uuid string) (domain.TaskCustomer, error)
	CreateFollowUp(ctx context.Context, input domain.PersistFollowUpInput) (domain.FollowUp, error)
	ListFollowUps(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	GetFollowUp(ctx context.Context, uuid string) (domain.FollowUp, error)
}

type ImageStorage interface {
	SaveFollowUpImages(ctx context.Context, followUpUUID string, images []domain.ImageUpload) ([]domain.StoredImage, error)
	DeleteImages(ctx context.Context, images []domain.StoredImage) error
}

type Service struct {
	repository Repository
	storage    ImageStorage
}

const (
	noteMaxLength       = 150
	imageMaxCount       = 3
	imageMaxTotalSize   = 5 * 1024 * 1024
	meetPersonNameLimit = 50
	meetPersonPhoneMax  = 20
	meetPersonEmailMax  = 100
)

var agreementFailureReasonOptions = map[string]struct{}{
	"Fiyat yüksek":                {},
	"Mesafe Uzak":                 {},
	"Bayi ile yaşanan sorunlar":   {},
	"Ekpertize ihtiyaç duymuyor":  {},
	"Kendisi yapıyor":             {},
	"Başka ekspertize yaptırıyor": {},
	"Değerlendirme":               {},
}

var visitTypeOptions = map[string]struct{}{
	"Yerinde Ziyaret": {},
}

var meetPersonTitleOptions = map[string]struct{}{
	"Genel Müdür":      {},
	"Satış Müdürü":     {},
	"Operasyon Müdürü": {},
	"Pazarlama Müdürü": {},
	"İşletme Müdürü":   {},
	"Bölge Müdürü":     {},
	"Şube Müdürü":      {},
	"Yönetici":         {},
	"Sahibi":           {},
	"Ortağı":           {},
}

var allowedImageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
}

func NewService(repository Repository, storage ImageStorage) *Service {
	return &Service{repository: repository, storage: storage}
}

func (s *Service) ListFollowUps(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.repository == nil {
		return domain.ListResult{}, ErrFollowUpListUnavailable
	}

	result, err := s.repository.ListFollowUps(ctx, normalizeListQuery(query))
	if err != nil {
		return domain.ListResult{}, ErrFollowUpListUnavailable
	}

	return result, nil
}

func (s *Service) GetFollowUp(ctx context.Context, followUpUUID string) (domain.FollowUp, error) {
	normalizedUUID := strings.TrimSpace(followUpUUID)
	if s == nil || s.repository == nil || normalizedUUID == "" {
		return domain.FollowUp{}, ErrFollowUpDetailUnavailable
	}

	followUp, err := s.repository.GetFollowUp(ctx, normalizedUUID)
	if err != nil {
		return domain.FollowUp{}, ErrFollowUpDetailUnavailable
	}

	return followUp, nil
}

func (s *Service) CreateFollowUp(ctx context.Context, input domain.CreateFollowUpInput) (domain.FollowUp, ValidationErrors, error) {
	normalizedInput := normalizeCreateFollowUpInput(input)
	validationErrors := validateCreateFollowUpInput(normalizedInput)
	if len(validationErrors) > 0 {
		return domain.FollowUp{}, validationErrors, ErrInvalidFollowUpCreateInput
	}

	if s == nil || s.repository == nil || s.storage == nil {
		return domain.FollowUp{}, nil, ErrFollowUpCreateUnavailable
	}

	taskCustomer, err := s.repository.FindTaskCustomerByUUID(ctx, normalizedInput.TasksCustomerUUID)
	if err != nil {
		return domain.FollowUp{}, ValidationErrors{
			"tasks_customer_uuid": "Seçili görev müşterisi bulunamadı.",
		}, ErrInvalidFollowUpCreateInput
	}
	if !canCreateFollowUpForTaskCustomer(taskCustomer.Status) {
		return domain.FollowUp{}, ValidationErrors{
			"tasks_customer_uuid": "Takip kaydı sadece bekleyen veya devam eden görev müşterileri için oluşturulabilir.",
		}, ErrInvalidFollowUpCreateInput
	}
	if taskCustomer.AssignedUserID != normalizedInput.AuthenticatedUserID {
		return domain.FollowUp{}, ValidationErrors{
			"tasks_customer_uuid": "Bu görev müşterisi için takip kaydı oluşturma yetkiniz yok.",
		}, ErrInvalidFollowUpCreateInput
	}

	followUpUUID := uuid.NewString()
	storedImages, err := s.storage.SaveFollowUpImages(ctx, followUpUUID, normalizedInput.Images)
	if err != nil {
		return domain.FollowUp{}, nil, ErrFollowUpCreateUnavailable
	}

	followUp, err := s.repository.CreateFollowUp(ctx, domain.PersistFollowUpInput{
		UUID:                   followUpUUID,
		TasksCustomerID:        taskCustomer.ID,
		TasksCustomerUUID:      taskCustomer.UUID,
		VisitType:              normalizedInput.VisitType,
		VisitDate:              normalizedInput.VisitDate,
		NextVisitDate:          normalizedInput.NextVisitDate,
		AgreementReached:       *normalizedInput.AgreementReached,
		AgreementFailureReason: normalizedInput.AgreementFailureReason,
		Note:                   normalizedInput.Note,
		Images:                 storedImages,
		MeetPeople:             normalizedInput.MeetPeople,
	})
	if err != nil {
		_ = s.storage.DeleteImages(ctx, storedImages)
		return domain.FollowUp{}, nil, ErrFollowUpCreateUnavailable
	}

	return followUp, nil, nil
}

func normalizeListQuery(query domain.ListQuery) domain.ListQuery {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	switch sortBy {
	case "visit_date", "next_visit_date", "agreement_reached":
	default:
		sortBy = ""
	}

	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	return domain.ListQuery{
		Page:                 query.Page,
		PerPage:              query.PerPage,
		Title:                strings.TrimSpace(query.Title),
		Customer:             strings.TrimSpace(query.Customer),
		AssignedUserID:       query.AssignedUserID,
		AssignedUserFullName: strings.TrimSpace(query.AssignedUserFullName),
		BranchName:           strings.TrimSpace(query.BranchName),
		VisitDate:            strings.TrimSpace(query.VisitDate),
		NextVisitDate:        strings.TrimSpace(query.NextVisitDate),
		SortBy:               sortBy,
		SortOrder:            sortOrder,
	}
}

func normalizeCreateFollowUpInput(input domain.CreateFollowUpInput) domain.CreateFollowUpInput {
	meetPeople := make([]domain.MeetPersonInput, 0, len(input.MeetPeople))
	for _, person := range input.MeetPeople {
		meetPeople = append(meetPeople, domain.MeetPersonInput{
			Title:   strings.TrimSpace(person.Title),
			Name:    strings.TrimSpace(person.Name),
			Surname: strings.TrimSpace(person.Surname),
			Phone:   strings.TrimSpace(person.Phone),
			Email:   strings.TrimSpace(person.Email),
		})
	}

	images := make([]domain.ImageUpload, 0, len(input.Images))
	for _, image := range input.Images {
		images = append(images, domain.ImageUpload{
			FileName:    strings.TrimSpace(image.FileName),
			ContentType: strings.ToLower(strings.TrimSpace(image.ContentType)),
			Size:        image.Size,
			Content:     image.Content,
		})
	}

	return domain.CreateFollowUpInput{
		AuthenticatedUserID:    input.AuthenticatedUserID,
		TasksCustomerUUID:      strings.TrimSpace(input.TasksCustomerUUID),
		VisitType:              strings.TrimSpace(input.VisitType),
		VisitDate:              strings.TrimSpace(input.VisitDate),
		NextVisitDate:          strings.TrimSpace(input.NextVisitDate),
		AgreementReached:       input.AgreementReached,
		AgreementFailureReason: strings.TrimSpace(input.AgreementFailureReason),
		Note:                   strings.TrimSpace(input.Note),
		Images:                 images,
		MeetPeople:             meetPeople,
	}
}

func validateCreateFollowUpInput(input domain.CreateFollowUpInput) ValidationErrors {
	errors := ValidationErrors{}

	requireField(errors, "tasks_customer_uuid", input.TasksCustomerUUID, "Görev müşterisi zorunludur.")
	requireField(errors, "visit_type", input.VisitType, "Ziyaret tipi zorunludur.")
	requireField(errors, "visit_date", input.VisitDate, "Ziyaret tarihi zorunludur.")
	if input.AuthenticatedUserID == 0 {
		errors["user"] = "Oturum kullanıcısı zorunludur."
	}
	validateVisitType(errors, input.VisitType)
	validateNote(errors, input.Note)
	validateAgreement(errors, input)
	validateDates(errors, input.VisitDate, input.NextVisitDate)
	validateImages(errors, input.Images)
	validateMeetPeople(errors, input.MeetPeople)

	return errors
}

func validateVisitType(errors ValidationErrors, visitType string) {
	if strings.TrimSpace(visitType) == "" {
		return
	}

	if _, ok := visitTypeOptions[visitType]; !ok {
		errors["visit_type"] = "Ziyaret tipi geçersiz."
	}
}

func requireField(errors ValidationErrors, field string, value string, message string) {
	if strings.TrimSpace(value) == "" {
		errors[field] = message
	}
}

func validateNote(errors ValidationErrors, note string) {
	if len([]rune(strings.TrimSpace(note))) > noteMaxLength {
		errors["note"] = "Not en fazla " + strconv.Itoa(noteMaxLength) + " karakter olabilir."
	}
}

func validateAgreement(errors ValidationErrors, input domain.CreateFollowUpInput) {
	if input.AgreementReached == nil {
		errors["agreement_reached"] = "Anlaşma durumu zorunludur."
		return
	}

	if *input.AgreementReached {
		if input.AgreementFailureReason != "" {
			errors["agreement_failure_reason"] = "Anlaşma sağlandıysa başarısızlık sebebi gönderilmemelidir."
		}
		return
	}

	if input.AgreementFailureReason == "" {
		errors["agreement_failure_reason"] = "Anlaşma sağlanmadıysa başarısızlık sebebi zorunludur."
		return
	}

	if _, ok := agreementFailureReasonOptions[input.AgreementFailureReason]; !ok {
		errors["agreement_failure_reason"] = "Başarısızlık sebebi geçersiz."
	}
}

func validateDates(errors ValidationErrors, visitDateValue string, nextVisitDateValue string) {
	visitDate, visitDateErr := parseDateTime(visitDateValue)
	if strings.TrimSpace(visitDateValue) != "" && visitDateErr != nil {
		errors["visit_date"] = "Ziyaret tarihi geçersiz."
	}

	nextVisitDate, nextVisitDateErr := parseDateTime(nextVisitDateValue)
	if strings.TrimSpace(nextVisitDateValue) != "" && nextVisitDateErr != nil {
		errors["next_visit_date"] = "Sonraki ziyaret tarihi geçersiz."
	}

	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if visitDateErr == nil && visitDate != nil && visitDate.Before(today) {
		errors["visit_date"] = "Ziyaret tarihi geçmiş bir tarih olamaz."
	}

	if visitDateErr == nil && nextVisitDateErr == nil && visitDate != nil && nextVisitDate != nil && nextVisitDate.Before(*visitDate) {
		errors["next_visit_date"] = "Sonraki ziyaret tarihi ziyaret tarihinden önce olamaz."
	}
}

func validateImages(errors ValidationErrors, images []domain.ImageUpload) {
	if len(images) > imageMaxCount {
		errors["images"] = "En fazla " + strconv.Itoa(imageMaxCount) + " dosya yüklenebilir."
		return
	}

	var totalSize int64
	for index, image := range images {
		field := "images." + strconv.Itoa(index)
		totalSize += image.Size
		if image.Size <= 0 || len(image.Content) == 0 {
			errors[field] = "Dosya boş olamaz."
			continue
		}

		contentType := image.ContentType
		if contentType == "" {
			contentType = strings.ToLower(http.DetectContentType(image.Content))
		}
		if _, ok := allowedImageContentTypes[contentType]; !ok {
			errors[field] = "Dosya tipi JPEG, PNG, GIF veya WebP olmalıdır."
		}

		if !hasAllowedImageExtension(image.FileName) {
			errors[field] = "Dosya uzantısı JPEG, PNG, GIF veya WebP olmalıdır."
		}
	}

	if totalSize > imageMaxTotalSize {
		errors["images"] = "Dosyaların toplam boyutu en fazla 5 MB olabilir."
	}
}

func validateMeetPeople(errors ValidationErrors, meetPeople []domain.MeetPersonInput) {
	if len(meetPeople) == 0 {
		errors["meet_people"] = "En az 1 kişi girilmelidir."
		return
	}

	for index, person := range meetPeople {
		prefix := "meet_people." + strconv.Itoa(index) + "."
		requireField(errors, prefix+"title", person.Title, "Ünvan zorunludur.")
		requireField(errors, prefix+"name", person.Name, "Ad zorunludur.")
		requireField(errors, prefix+"surname", person.Surname, "Soyad zorunludur.")
		requireField(errors, prefix+"phone", person.Phone, "Telefon zorunludur.")

		if person.Title != "" {
			if _, ok := meetPersonTitleOptions[person.Title]; !ok {
				errors[prefix+"title"] = "Ünvan geçersiz."
			}
		}
		validateMaxLength(errors, prefix+"name", person.Name, "Ad", meetPersonNameLimit)
		validateMaxLength(errors, prefix+"surname", person.Surname, "Soyad", meetPersonNameLimit)
		validateMaxLength(errors, prefix+"phone", person.Phone, "Telefon", meetPersonPhoneMax)
		validateMaxLength(errors, prefix+"email", person.Email, "E-posta", meetPersonEmailMax)
	}
}

func validateMaxLength(errors ValidationErrors, field string, value string, label string, limit int) {
	if len([]rune(strings.TrimSpace(value))) > limit {
		errors[field] = label + " en fazla " + strconv.Itoa(limit) + " karakter olabilir."
	}
}

func canCreateFollowUpForTaskCustomer(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending", "in_progress":
		return true
	default:
		return false
	}
}

func hasAllowedImageExtension(fileName string) bool {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(fileName))) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func parseDateTime(value string) (*time.Time, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var parseErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, trimmedValue, time.Local)
		if err == nil {
			return &parsed, nil
		}
		parseErr = err
	}

	return nil, parseErr
}
