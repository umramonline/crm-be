package greeting

import (
	domain "github.com/umran/new.crm/backend/internal/domain/greeting"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetHello() domain.Greeting {
	return domain.New()
}
