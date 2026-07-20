package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umran/new.crm/backend/internal/dashboard/domain"
)

type fakeStatsRepository struct {
	potentialCustomers int64
	totalCustomers     int64
	customerVisits     int64
	newCustomers       int64
	vehicleStock       int64
	taskStats          domain.TaskStats
	overdueTasks       int64
	err                error
}

func (f *fakeStatsRepository) CountPotentialCustomers(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.potentialCustomers, nil
}

func (f *fakeStatsRepository) CountTotalCustomers(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.totalCustomers, nil
}

func (f *fakeStatsRepository) CountCustomerVisits(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.customerVisits, nil
}

func (f *fakeStatsRepository) CountNewCustomers(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.newCustomers, nil
}

func (f *fakeStatsRepository) SumVehicleStock(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.vehicleStock, nil
}

func (f *fakeStatsRepository) CountTaskStats(_ context.Context, _ domain.Filter) (domain.TaskStats, error) {
	if f.err != nil {
		return domain.TaskStats{}, f.err
	}

	return f.taskStats, nil
}

func (f *fakeStatsRepository) CountOverdueTasks(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.overdueTasks, nil
}

type fakeUmramonlineStatsProvider struct {
	vehicleEntries int64
	totalAmount    float64
	loadedCredit   float64
	err            error
}

func (f *fakeUmramonlineStatsProvider) CountVehicleEntries(_ context.Context, _ domain.Filter) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.vehicleEntries, nil
}

func (f *fakeUmramonlineStatsProvider) SumTotalAmount(_ context.Context, _ domain.Filter) (float64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.totalAmount, nil
}

func (f *fakeUmramonlineStatsProvider) SumLoadedCredit(_ context.Context, _ domain.Filter) (float64, error) {
	if f.err != nil {
		return 0, f.err
	}

	return f.loadedCredit, nil
}

func validDashboardFilter() domain.Filter {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	return domain.Filter{
		StartDate:        startDate,
		EndDate:          endDate,
		BranchIDs:        []uint64{1, 2},
		AllowAllBranches: false,
	}
}

func TestGetDashboardAggregatesAllStatsInParallel(t *testing.T) {
	service := NewService(
		&fakeStatsRepository{
			potentialCustomers: 1,
			totalCustomers:     2,
			customerVisits:     3,
			newCustomers:       4,
			vehicleStock:       5,
			taskStats: domain.TaskStats{
				PendingCount:    6,
				InProgressCount: 7,
				CompletedCount:  8,
			},
			overdueTasks: 9,
		},
		&fakeUmramonlineStatsProvider{
			vehicleEntries: 10,
			totalAmount:    11.5,
			loadedCredit:   12.5,
		},
	)

	stats, validationErrors, err := service.GetDashboard(context.Background(), validDashboardFilter())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(validationErrors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", validationErrors)
	}

	if stats.PotentialCustomerCount != 1 ||
		stats.TotalCustomerCount != 2 ||
		stats.CustomerVisitCount != 3 ||
		stats.NewCustomerCount != 4 ||
		stats.VehicleStockCount != 5 ||
		stats.PendingTaskCount != 6 ||
		stats.InProgressTaskCount != 7 ||
		stats.CompletedTaskCount != 8 ||
		stats.OverdueTaskCount != 9 ||
		stats.VehicleEntryCount != 10 ||
		stats.TotalAmount != 11.5 ||
		stats.LoadedCreditAmount != 12.5 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

func TestGetDashboardRejectsInvalidDateRange(t *testing.T) {
	service := NewService(&fakeStatsRepository{}, &fakeUmramonlineStatsProvider{})

	filter := validDashboardFilter()
	filter.EndDate = filter.StartDate.AddDate(0, 0, -1)

	_, validationErrors, err := service.GetDashboard(context.Background(), filter)
	if !errors.Is(err, ErrInvalidDashboardFilter) {
		t.Fatalf("expected ErrInvalidDashboardFilter, got %v", err)
	}

	if validationErrors["end_date"] == "" {
		t.Fatal("expected end_date validation error")
	}
}

func TestGetDashboardReturnsUnavailableWhenRepositoryFails(t *testing.T) {
	service := NewService(
		&fakeStatsRepository{err: errors.New("db down")},
		&fakeUmramonlineStatsProvider{},
	)

	_, _, err := service.GetDashboard(context.Background(), validDashboardFilter())
	if !errors.Is(err, ErrDashboardUnavailable) {
		t.Fatalf("expected ErrDashboardUnavailable, got %v", err)
	}
}

func TestNormalizeFilterDates(t *testing.T) {
	startDate := time.Date(2026, 3, 10, 15, 30, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 20, 8, 15, 0, 0, time.UTC)

	filter := NormalizeFilterDates(startDate, endDate)
	if filter.StartDate.Hour() != 0 || filter.StartDate.Minute() != 0 {
		t.Fatalf("unexpected start date normalization: %#v", filter.StartDate)
	}

	if filter.EndDate.Hour() != 23 || filter.EndDate.Minute() != 59 || filter.EndDate.Second() != 59 {
		t.Fatalf("unexpected end date normalization: %#v", filter.EndDate)
	}
}
