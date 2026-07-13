package http

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	authapp "github.com/umran/new.crm/backend/internal/auth/application"
	"github.com/umran/new.crm/backend/internal/followup/application"
	"github.com/umran/new.crm/backend/internal/followup/domain"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

type createFollowUpRequest struct {
	TasksCustomerUUID      string                    `json:"tasks_customer_uuid"`
	VisitType              string                    `json:"visit_type"`
	VisitDate              string                    `json:"visit_date"`
	NextVisitDate          string                    `json:"next_visit_date"`
	AgreementReached       *bool                     `json:"agreement_reached"`
	AgreementFailureReason string                    `json:"agreement_failure_reason"`
	Note                   string                    `json:"note"`
	MeetPeople             []createMeetPersonRequest `json:"meet_people"`
}

type updateFollowUpRequest struct {
	VisitType              string                    `json:"visit_type"`
	NextVisitDate          string                    `json:"next_visit_date"`
	AgreementReached       *bool                     `json:"agreement_reached"`
	AgreementFailureReason string                    `json:"agreement_failure_reason"`
	Note                   string                    `json:"note"`
	ExistingImageUUIDs     []string                  `json:"existing_image_uuids"`
	MeetPeople             []createMeetPersonRequest `json:"meet_people"`
}

type createMeetPersonRequest struct {
	Title   string `json:"title"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/follow-ups", authRequired, h.ListFollowUps)
	router.Get("/follow-ups/assigned-to-me", authRequired, h.ListAssignedFollowUps)
	router.Get("/follow-ups/:uuid", authRequired, h.GetFollowUp)
	router.Post("/follow-ups", authRequired, h.CreateFollowUp)
	router.Put("/follow-ups/:uuid", authRequired, h.UpdateFollowUp)
}

func (h *Handler) ListFollowUps(c *fiber.Ctx) error {
	result, err := h.service.ListFollowUps(c.UserContext(), domain.ListQuery{
		Page:                 queryInt(c, "page", 1),
		PerPage:              queryInt(c, "per_page", 10),
		Title:                c.Query("title"),
		Customer:             c.Query("customer"),
		AssignedUserFullName: c.Query("assigned_user_full_name"),
		BranchName:           c.Query("branch_name"),
		VisitDate:            c.Query("visit_date"),
		NextVisitDate:        c.Query("next_visit_date"),
		SortBy:               c.Query("sort_by"),
		SortOrder:            c.Query("sort_order"),
	})
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Takip kayıtları şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Takip kayıtları getirildi.", result)
}

func (h *Handler) ListAssignedFollowUps(c *fiber.Ctx) error {
	claims := c.Locals("claims").(authapp.SessionTokenClaims)
	result, err := h.service.ListFollowUps(c.UserContext(), domain.ListQuery{
		Page:                 queryInt(c, "page", 1),
		PerPage:              queryInt(c, "per_page", 10),
		Title:                c.Query("title"),
		Customer:             c.Query("customer"),
		AssignedUserID:       claims.UserId,
		AssignedUserFullName: c.Query("assigned_user_full_name"),
		BranchName:           c.Query("branch_name"),
		VisitDate:            c.Query("visit_date"),
		NextVisitDate:        c.Query("next_visit_date"),
		SortBy:               c.Query("sort_by"),
		SortOrder:            c.Query("sort_order"),
	})
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Takip kayıtları şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Takip kayıtları getirildi.", result)
}

func (h *Handler) GetFollowUp(c *fiber.Ctx) error {
	followUp, err := h.service.GetFollowUp(c.UserContext(), c.Params("uuid"))
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Takip kaydı detayı şu anda getirilemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Takip kaydı detayı getirildi.", followUp)
}

func (h *Handler) CreateFollowUp(c *fiber.Ctx) error {
	input, validationErrors, err := createFollowUpInput(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Takip kaydı bilgileri geçersiz.", validationErrors)
	}
	claims := c.Locals("claims").(authapp.SessionTokenClaims)
	input.AuthenticatedUserID = claims.UserId

	followUp, validationErrors, err := h.service.CreateFollowUp(c.UserContext(), input)
	if err != nil {
		if err == application.ErrInvalidFollowUpCreateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Takip kaydı bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Takip kaydı şu anda oluşturulamadı.", nil)
	}

	return response.Success(c, fiber.StatusCreated, "Takip kaydı oluşturuldu.", followUp)
}

func (h *Handler) UpdateFollowUp(c *fiber.Ctx) error {
	input, validationErrors, err := updateFollowUpInput(c)
	if err != nil {
		if err == fiber.ErrUnsupportedMediaType {
			return response.Error(c, fiber.StatusUnsupportedMediaType, "Takip kaydı multipart/form-data olarak gönderilmelidir.", validationErrors)
		}

		return response.Error(c, fiber.StatusUnprocessableEntity, "Takip kaydı bilgileri geçersiz.", validationErrors)
	}
	input.UUID = c.Params("uuid")

	followUp, validationErrors, err := h.service.UpdateFollowUp(c.UserContext(), input)
	if err != nil {
		if err == application.ErrInvalidFollowUpUpdateInput {
			return response.Error(c, fiber.StatusUnprocessableEntity, "Takip kaydı bilgileri geçersiz.", validationErrors)
		}

		return response.Error(c, fiber.StatusServiceUnavailable, "Takip kaydı şu anda güncellenemedi.", nil)
	}

	return response.Success(c, fiber.StatusOK, "Takip kaydı güncellendi.", followUp)
}

func createFollowUpInput(c *fiber.Ctx) (domain.CreateFollowUpInput, application.ValidationErrors, error) {
	contentType := strings.ToLower(c.Get(fiber.HeaderContentType))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return multipartFollowUpInput(c)
	}

	return domain.CreateFollowUpInput{}, application.ValidationErrors{}, nil
}

func updateFollowUpInput(c *fiber.Ctx) (domain.UpdateFollowUpInput, application.ValidationErrors, error) {
	contentType := strings.ToLower(c.Get(fiber.HeaderContentType))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return multipartUpdateFollowUpInput(c)
	}

	return domain.UpdateFollowUpInput{}, application.ValidationErrors{
		"request": "Takip kaydı multipart/form-data olarak gönderilmelidir.",
	}, fiber.ErrUnsupportedMediaType
}

func multipartFollowUpInput(c *fiber.Ctx) (domain.CreateFollowUpInput, application.ValidationErrors, error) {
	agreementReached, err := parseBoolPointer(c.FormValue("agreement_reached"))
	if err != nil {
		return domain.CreateFollowUpInput{}, application.ValidationErrors{
			"agreement_reached": "Anlaşma durumu boolean olmalıdır.",
		}, err
	}

	meetPeople, err := parseMeetPeople(c.FormValue("meet_people"))
	if err != nil {
		return domain.CreateFollowUpInput{}, application.ValidationErrors{
			"meet_people": "Görüşülen kişiler geçersiz.",
		}, err
	}

	images, err := multipartImages(c)
	if err != nil {
		return domain.CreateFollowUpInput{}, application.ValidationErrors{
			"images": "Dosyalar okunamadı.",
		}, err
	}

	return requestToInput(createFollowUpRequest{
		TasksCustomerUUID:      c.FormValue("tasks_customer_uuid"),
		VisitType:              c.FormValue("visit_type"),
		VisitDate:              c.FormValue("visit_date"),
		NextVisitDate:          c.FormValue("next_visit_date"),
		AgreementReached:       agreementReached,
		AgreementFailureReason: c.FormValue("agreement_failure_reason"),
		Note:                   c.FormValue("note"),
		MeetPeople:             meetPeople,
	}, images), nil, nil
}

func multipartUpdateFollowUpInput(c *fiber.Ctx) (domain.UpdateFollowUpInput, application.ValidationErrors, error) {
	agreementReached, err := parseBoolPointer(c.FormValue("agreement_reached"))
	if err != nil {
		return domain.UpdateFollowUpInput{}, application.ValidationErrors{
			"agreement_reached": "Anlaşma durumu boolean olmalıdır.",
		}, err
	}

	meetPeople, err := parseMeetPeople(c.FormValue("meet_people"))
	if err != nil {
		return domain.UpdateFollowUpInput{}, application.ValidationErrors{
			"meet_people": "Görüşülen kişiler geçersiz.",
		}, err
	}

	existingImageUUIDs, err := parseStringArray(c.FormValue("existing_image_uuids"))
	if err != nil {
		return domain.UpdateFollowUpInput{}, application.ValidationErrors{
			"existing_image_uuids": "Mevcut resim bilgileri geçersiz.",
		}, err
	}

	images, err := multipartImages(c)
	if err != nil {
		return domain.UpdateFollowUpInput{}, application.ValidationErrors{
			"images": "Dosyalar okunamadı.",
		}, err
	}

	return updateRequestToInput(updateFollowUpRequest{
		VisitType:              c.FormValue("visit_type"),
		NextVisitDate:          c.FormValue("next_visit_date"),
		AgreementReached:       agreementReached,
		AgreementFailureReason: c.FormValue("agreement_failure_reason"),
		Note:                   c.FormValue("note"),
		ExistingImageUUIDs:     existingImageUUIDs,
		MeetPeople:             meetPeople,
	}, images), nil, nil
}

func requestToInput(request createFollowUpRequest, images []domain.ImageUpload) domain.CreateFollowUpInput {
	meetPeople := make([]domain.MeetPersonInput, 0, len(request.MeetPeople))
	for _, person := range request.MeetPeople {
		meetPeople = append(meetPeople, domain.MeetPersonInput{
			Title:   person.Title,
			Name:    person.Name,
			Surname: person.Surname,
			Phone:   person.Phone,
			Email:   person.Email,
		})
	}

	return domain.CreateFollowUpInput{
		TasksCustomerUUID:      request.TasksCustomerUUID,
		VisitType:              request.VisitType,
		VisitDate:              request.VisitDate,
		NextVisitDate:          request.NextVisitDate,
		AgreementReached:       request.AgreementReached,
		AgreementFailureReason: request.AgreementFailureReason,
		Note:                   request.Note,
		Images:                 images,
		MeetPeople:             meetPeople,
	}
}

func updateRequestToInput(request updateFollowUpRequest, images []domain.ImageUpload) domain.UpdateFollowUpInput {
	meetPeople := make([]domain.MeetPersonInput, 0, len(request.MeetPeople))
	for _, person := range request.MeetPeople {
		meetPeople = append(meetPeople, domain.MeetPersonInput{
			Title:   person.Title,
			Name:    person.Name,
			Surname: person.Surname,
			Phone:   person.Phone,
			Email:   person.Email,
		})
	}

	return domain.UpdateFollowUpInput{
		VisitType:              request.VisitType,
		NextVisitDate:          request.NextVisitDate,
		AgreementReached:       request.AgreementReached,
		AgreementFailureReason: request.AgreementFailureReason,
		Note:                   request.Note,
		Images:                 images,
		ExistingImageUUIDs:     request.ExistingImageUUIDs,
		MeetPeople:             meetPeople,
	}
}

func parseBoolPointer(value string) (*bool, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseBool(trimmedValue)
	if err != nil {
		return nil, err
	}

	return &parsedValue, nil
}

func parseMeetPeople(value string) ([]createMeetPersonRequest, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	var meetPeople []createMeetPersonRequest
	if err := json.Unmarshal([]byte(trimmedValue), &meetPeople); err != nil {
		return nil, err
	}

	return meetPeople, nil
}

func parseStringArray(value string) ([]string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, nil
	}

	var values []string
	if err := json.Unmarshal([]byte(trimmedValue), &values); err != nil {
		return nil, err
	}

	return values, nil
}

func multipartImages(c *fiber.Ctx) ([]domain.ImageUpload, error) {
	form, err := c.MultipartForm()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "request content-type isn't multipart/form-data") {
			return nil, nil
		}
		return nil, err
	}
	if form == nil || form.File == nil {
		return nil, nil
	}

	fileHeaders := append([]*multipart.FileHeader{}, form.File["images"]...)
	fileHeaders = append(fileHeaders, form.File["images[]"]...)
	images := make([]domain.ImageUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		image, err := imageUploadFromFileHeader(fileHeader)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func imageUploadFromFileHeader(fileHeader *multipart.FileHeader) (domain.ImageUpload, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return domain.ImageUpload{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return domain.ImageUpload{}, err
	}

	return domain.ImageUpload{
		FileName:    fileHeader.Filename,
		ContentType: headerContentType(fileHeader),
		Size:        fileHeader.Size,
		Content:     content,
	}, nil
}

func headerContentType(fileHeader *multipart.FileHeader) string {
	if fileHeader == nil || fileHeader.Header == nil {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(strings.Split(fileHeader.Header.Get("Content-Type"), ";")[0]))
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
