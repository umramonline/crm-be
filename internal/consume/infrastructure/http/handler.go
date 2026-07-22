package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/consume/application"
	"github.com/umran/new.crm/backend/internal/consume/domain"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, apiKeyRequired fiber.Handler) {
	router.Post("/consume", apiKeyRequired, h.Consume)
}

type consumeRequest struct {
	EventID          string              `json:"event_id"`
	EventType        string              `json:"event_type"`
	UOId             uint64              `json:"uo_id"`
	BranchID         int32               `json:"branch_id"`
	Unvan            string              `json:"unvan"`
	Ad               string              `json:"ad"`
	Soyad            string              `json:"soyad"`
	YetkiliAdi       string              `json:"yetkili_adi"`
	Cep              string              `json:"cep"`
	Telefon          string              `json:"telefon"`
	Fax              string              `json:"fax"`
	Eposta           string              `json:"eposta"`
	Web              string              `json:"web"`
	Mahalle          string              `json:"mahalle"`
	Cadde            string              `json:"cadde"`
	Sokak            string              `json:"sokak"`
	Semt             string              `json:"semt"`
	KapiNo           string              `json:"kapi_no"`
	IlKodu           string              `json:"il_kodu"`
	IlceKodu         string              `json:"ilce_kodu"`
	Ulke             string              `json:"ulke"`
	DogumTarihi      *string             `json:"dogum_tarihi"`
	VadeGunu         *string             `json:"vade_gunu"`
	VergiDairesi     string              `json:"vergi_dairesi"`
	VergiDairesiKodu string              `json:"vergi_dairesi_kodu"`
	VergiNo          string              `json:"vergi_no"`
	TCNo             string              `json:"tc_no"`
	Type             string              `json:"type"`
	Mersis           string              `json:"mersis"`
	PasaportNo       string              `json:"pasaport_no"`
	PasaportBelge    string              `json:"pasaport_belge"`
	EsbisNo          string              `json:"esbis_no"`
	YetkiBelgeNo     string              `json:"yetki_belge_no"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
	Telephones       []telephoneRequest  `json:"telephones"`
	OccurredAt       string              `json:"occurred_at"`
}

type telephoneRequest struct {
	PhoneNumber string `json:"phone_number"`
	Title       string `json:"title"`
}

func (h *Handler) Consume(c *fiber.Ctx) error {
	var request consumeRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", fiber.Map{
			"body": "JSON parse edilemedi.",
		})
	}

	result, err := h.service.Consume(c.UserContext(), toCustomerCreatedEvent(request))
	if err != nil {
		switch err {
		case application.ErrInvalidEventPayload:
			return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz event payload.", fiber.Map{
				"event_id":  "event_id zorunludur.",
				"uo_id":     "uo_id zorunludur.",
				"event_type": "event_type zorunludur.",
			})
		case application.ErrUnsupportedEventType:
			return response.Error(c, fiber.StatusUnprocessableEntity, "Desteklenmeyen event_type.", fiber.Map{
				"event_type": "Şu an yalnızca customer.created desteklenmektedir.",
			})
		default:
			return response.Error(c, fiber.StatusInternalServerError, "Event işlenemedi.", nil)
		}
	}

	status := fiber.StatusOK
	message := "Event consumed."
	if result.Action == "created" {
		status = fiber.StatusCreated
		message = "Customer created."
	} else if result.Action == "updated" {
		message = "Customer updated."
	} else if result.Action == "already_processed" {
		message = "Event already processed."
	}

	return response.Success(c, status, message, result)
}

func toCustomerCreatedEvent(request consumeRequest) domain.CustomerCreatedEvent {
	telephones := make([]domain.Telephone, 0, len(request.Telephones))
	for _, telephone := range request.Telephones {
		telephones = append(telephones, domain.Telephone{
			PhoneNumber: telephone.PhoneNumber,
			Title:       telephone.Title,
		})
	}

	return domain.CustomerCreatedEvent{
		EventID:          strings.TrimSpace(request.EventID),
		EventType:        strings.TrimSpace(request.EventType),
		UOId:             request.UOId,
		BranchID:         request.BranchID,
		Unvan:            request.Unvan,
		Ad:               request.Ad,
		Soyad:            request.Soyad,
		YetkiliAdi:       request.YetkiliAdi,
		Cep:              request.Cep,
		Telefon:          request.Telefon,
		Fax:              request.Fax,
		Eposta:           request.Eposta,
		Web:              request.Web,
		Mahalle:          request.Mahalle,
		Cadde:            request.Cadde,
		Sokak:            request.Sokak,
		Semt:             request.Semt,
		KapiNo:           request.KapiNo,
		IlKodu:           request.IlKodu,
		IlceKodu:         request.IlceKodu,
		Ulke:             request.Ulke,
		DogumTarihi:      request.DogumTarihi,
		VadeGunu:         request.VadeGunu,
		VergiDairesi:     request.VergiDairesi,
		VergiDairesiKodu: request.VergiDairesiKodu,
		VergiNo:          request.VergiNo,
		TCNo:             request.TCNo,
		Type:             request.Type,
		Mersis:           request.Mersis,
		PasaportNo:       request.PasaportNo,
		PasaportBelge:    request.PasaportBelge,
		EsbisNo:          request.EsbisNo,
		YetkiBelgeNo:     request.YetkiBelgeNo,
		CreatedAt:        request.CreatedAt,
		UpdatedAt:        request.UpdatedAt,
		Telephones:       telephones,
		OccurredAt:       request.OccurredAt,
	}
}
