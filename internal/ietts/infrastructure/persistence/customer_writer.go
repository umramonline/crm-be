package persistence

import (
	"context"

	customerpersistence "github.com/umran/new.crm/backend/internal/customer/infrastructure/persistence"
	"github.com/umran/new.crm/backend/internal/ietts/domain"
)

type CustomerWriter struct {
	repository *customerpersistence.Repository
}

func NewCustomerWriter(repository *customerpersistence.Repository) *CustomerWriter {
	return &CustomerWriter{repository: repository}
}

func (w *CustomerWriter) CreateCustomerFromIetts(
	ctx context.Context,
	input domain.CustomerFromIettsInput,
) (uint64, error) {
	if w == nil || w.repository == nil {
		return 0, ErrCustomerWriterUnavailable
	}

	return w.repository.CreateCustomerFromIetts(
		ctx,
		input.Unvan,
		input.Ad,
		input.Soyad,
		input.AddressDetail,
	)
}
