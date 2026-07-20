package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/umran/new.crm/backend/internal/dashboard/domain"
)

var ErrDashboardUnavailable = errors.New("dashboard unavailable")

var ErrInvalidDashboardFilter = errors.New("invalid dashboard filter")

type ValidationErrors map[string]string

type StatsRepository interface {
	CountPotentialCustomers(ctx context.Context, filter domain.Filter) (int64, error)
	CountTotalCustomers(ctx context.Context, filter domain.Filter) (int64, error)
	CountCustomerVisits(ctx context.Context, filter domain.Filter) (int64, error)
	CountNewCustomers(ctx context.Context, filter domain.Filter) (int64, error)
	SumVehicleStock(ctx context.Context, filter domain.Filter) (int64, error)
	CountTaskStats(ctx context.Context, filter domain.Filter) (domain.TaskStats, error)
	CountOverdueTasks(ctx context.Context, filter domain.Filter) (int64, error)
}

type UmramonlineStatsProvider interface {
	CountVehicleEntries(ctx context.Context, filter domain.Filter) (int64, error)
	SumTotalAmount(ctx context.Context, filter domain.Filter) (float64, error)
	SumLoadedCredit(ctx context.Context, filter domain.Filter) (float64, error)
}

type Service struct {
	repository StatsRepository
	provider   UmramonlineStatsProvider
}

func NewService(repository StatsRepository, provider UmramonlineStatsProvider) *Service {
	return &Service{
		repository: repository,
		provider:   provider,
	}
}

func (s *Service) GetDashboard(ctx context.Context, filter domain.Filter) (domain.Stats, ValidationErrors, error) {
	if s == nil || s.repository == nil || s.provider == nil {
		return domain.Stats{}, nil, ErrDashboardUnavailable
	}

	validationErrors := validateFilter(filter)
	if len(validationErrors) > 0 {
		return domain.Stats{}, validationErrors, ErrInvalidDashboardFilter
	}

	var (
		stats domain.Stats
		mu    sync.Mutex
	)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		value, err := s.repository.CountPotentialCustomers(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.PotentialCustomerCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.CountTotalCustomers(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.TotalCustomerCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.CountCustomerVisits(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.CustomerVisitCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.CountNewCustomers(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.NewCustomerCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.provider.CountVehicleEntries(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.VehicleEntryCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.provider.SumTotalAmount(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.TotalAmount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.provider.SumLoadedCredit(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.LoadedCreditAmount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.SumVehicleStock(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.VehicleStockCount = value
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.CountTaskStats(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.PendingTaskCount = value.PendingCount
		stats.InProgressTaskCount = value.InProgressCount
		stats.CompletedTaskCount = value.CompletedCount
		mu.Unlock()
		return nil
	})

	group.Go(func() error {
		value, err := s.repository.CountOverdueTasks(groupCtx, filter)
		if err != nil {
			return err
		}
		mu.Lock()
		stats.OverdueTaskCount = value
		mu.Unlock()
		return nil
	})

	if err := group.Wait(); err != nil {
		return domain.Stats{}, nil, ErrDashboardUnavailable
	}

	return stats, nil, nil
}

func validateFilter(filter domain.Filter) ValidationErrors {
	errors := ValidationErrors{}

	if filter.StartDate.IsZero() {
		errors["start_date"] = "Başlangıç tarihi zorunludur."
	}

	if filter.EndDate.IsZero() {
		errors["end_date"] = "Bitiş tarihi zorunludur."
	}

	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() && filter.EndDate.Before(filter.StartDate) {
		errors["end_date"] = "Bitiş tarihi başlangıç tarihinden önce olamaz."
	}

	return errors
}

func NormalizeFilterDates(startDate time.Time, endDate time.Time) domain.Filter {
	normalizedStart := time.Date(
		startDate.Year(), startDate.Month(), startDate.Day(),
		0, 0, 0, 0, startDate.Location(),
	)
	normalizedEnd := time.Date(
		endDate.Year(), endDate.Month(), endDate.Day(),
		23, 59, 59, 0, endDate.Location(),
	)

	return domain.Filter{
		StartDate: normalizedStart,
		EndDate:   normalizedEnd,
	}
}
