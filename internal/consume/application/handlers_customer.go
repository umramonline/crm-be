package application

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/umran/new.crm/backend/internal/consume/domain"
)

func (s *Service) handleCustomerCreated(ctx context.Context, command domain.ConsumeCommand) (domain.ConsumeResult, error) {
	event, err := decodeCustomerEvent(command)
	if err != nil {
		return domain.ConsumeResult{}, err
	}

	return s.repository.ConsumeCustomerCreated(ctx, event)
}

func (s *Service) handleCustomerUpdated(ctx context.Context, command domain.ConsumeCommand) (domain.ConsumeResult, error) {
	event, err := decodeCustomerEvent(command)
	if err != nil {
		return domain.ConsumeResult{}, err
	}

	return s.repository.ConsumeCustomerUpdated(ctx, event)
}

func decodeCustomerEvent(command domain.ConsumeCommand) (domain.CustomerEvent, error) {
	var payload customerEventPayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		return domain.CustomerEvent{}, ErrInvalidEventPayload
	}

	event := mapCustomerEventPayload(command.EventID, command.EventType, payload)
	if event.UOId == 0 {
		return domain.CustomerEvent{}, ErrInvalidEventPayload
	}

	if strings.TrimSpace(event.OccurredAt) == "" {
		return domain.CustomerEvent{}, ErrInvalidEventPayload
	}

	return event, nil
}

func mapCustomerEventPayload(eventID, eventType string, payload customerEventPayload) domain.CustomerEvent {
	telephones := make([]domain.Telephone, 0, len(payload.Telephones))
	for _, telephone := range payload.Telephones {
		telephones = append(telephones, domain.Telephone{
			PhoneNumber: telephone.PhoneNumber,
			Title:       telephone.Title,
		})
	}

	return domain.CustomerEvent{
		EventID:          eventID,
		EventType:        eventType,
		UOId:             payload.UOId,
		BranchID:         payload.BranchID,
		Unvan:            payload.Unvan,
		Ad:               payload.Ad,
		Soyad:            payload.Soyad,
		YetkiliAdi:       payload.YetkiliAdi,
		Cep:              payload.Cep,
		Telefon:          payload.Telefon,
		Fax:              payload.Fax,
		Eposta:           payload.Eposta,
		Web:              payload.Web,
		Mahalle:          payload.Mahalle,
		Cadde:            payload.Cadde,
		Sokak:            payload.Sokak,
		Semt:             payload.Semt,
		KapiNo:           payload.KapiNo,
		IlKodu:           payload.IlKodu,
		IlceKodu:         payload.IlceKodu,
		Ulke:             payload.Ulke,
		DogumTarihi:      payload.DogumTarihi,
		VadeGunu:         payload.VadeGunu,
		VergiDairesi:     payload.VergiDairesi,
		VergiDairesiKodu: payload.VergiDairesiKodu,
		VergiNo:          payload.VergiNo,
		TCNo:             payload.TCNo,
		Type:             payload.Type,
		Mersis:           payload.Mersis,
		PasaportNo:       payload.PasaportNo,
		PasaportBelge:    payload.PasaportBelge,
		EsbisNo:          payload.EsbisNo,
		YetkiBelgeNo:     payload.YetkiBelgeNo,
		CreatedAt:        payload.CreatedAt,
		UpdatedAt:        payload.UpdatedAt,
		Telephones:       telephones,
		OccurredAt:       payload.OccurredAt,
	}
}
