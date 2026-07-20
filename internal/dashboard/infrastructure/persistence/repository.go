package persistence

import (
	"context"
	"time"

	"github.com/umran/new.crm/backend/internal/dashboard/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountPotentialCustomers(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	query := r.db.WithContext(ctx).
		Model(&customerModel{}).
		Where("uo_id = ?", 0)

	query = applyBranchFilter(query, filter)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) CountTotalCustomers(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	query := applyBranchFilter(r.db.WithContext(ctx).Model(&customerModel{}), filter)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) CountCustomerVisits(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	query := r.db.WithContext(ctx).
		Table("tasks_follow_ups tfu").
		Joins("JOIN tasks_customers tc ON tc.id = tfu.tasks_customer_id").
		Joins("JOIN customers c ON c.id = tc.customer_id AND c.deleted_at IS NULL").
		Where("tfu.deleted_at IS NULL").
		Where("tfu.created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate)

	query = applyCustomerBranchFilter(query, filter, "c.branch_id")

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) CountNewCustomers(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	windowEnd := filter.StartDate.AddDate(0, 0, 7)
	query := r.db.WithContext(ctx).
		Model(&customerModel{}).
		Where("created_at >= ? AND created_at < ?", filter.StartDate, windowEnd)

	query = applyBranchFilter(query, filter)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) SumVehicleStock(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	query := applyBranchFilter(r.db.WithContext(ctx).Model(&customerModel{}), filter)

	var total int64
	if err := query.Select("COALESCE(SUM(vehicle_stock_count), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) CountTaskStats(ctx context.Context, filter domain.Filter) (domain.TaskStats, error) {
	if r == nil || r.db == nil {
		return domain.TaskStats{}, gorm.ErrInvalidDB
	}

	type taskStatsRow struct {
		PendingCount    int64
		InProgressCount int64
		CompletedCount  int64
	}

	query := r.db.WithContext(ctx).
		Table("tasks_customers tc").
		Joins("JOIN customers c ON c.id = tc.customer_id AND c.deleted_at IS NULL").
		Where("tc.created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate)

	query = applyCustomerBranchFilter(query, filter, "c.branch_id")

	var row taskStatsRow
	if err := query.Select(`
		COALESCE(SUM(CASE WHEN tc.status = 'pending' THEN 1 ELSE 0 END), 0) AS pending_count,
		COALESCE(SUM(CASE WHEN tc.status = 'in_progress' THEN 1 ELSE 0 END), 0) AS in_progress_count,
		COALESCE(SUM(CASE WHEN tc.status = 'completed' THEN 1 ELSE 0 END), 0) AS completed_count
	`).Scan(&row).Error; err != nil {
		return domain.TaskStats{}, err
	}

	return domain.TaskStats{
		PendingCount:    row.PendingCount,
		InProgressCount: row.InProgressCount,
		CompletedCount:  row.CompletedCount,
	}, nil
}

func (r *Repository) CountOverdueTasks(ctx context.Context, filter domain.Filter) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}

	query := r.db.WithContext(ctx).
		Table("tasks_customers tc").
		Joins("JOIN tasks t ON t.id = tc.task_id AND t.deleted_at IS NULL").
		Joins("JOIN customers c ON c.id = tc.customer_id AND c.deleted_at IS NULL").
		Where(`
			(
				(t.due_date IS NULL AND t.visit_date < CURDATE())
			 OR (t.due_date IS NOT NULL AND t.due_date < CURDATE())
			)
		`).
		Where("tc.status = ?", "pending")

	query = applyCustomerBranchFilter(query, filter, "c.branch_id")

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

type customerModel struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement"`
	UOId              uint64         `gorm:"column:uo_id;type:bigint"`
	BranchID          *int32         `gorm:"column:branch_id;type:int"`
	VehicleStockCount *int32         `gorm:"column:vehicle_stock_count;type:int"`
	CreatedAt         time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (customerModel) TableName() string {
	return "customers"
}

func applyBranchFilter(query *gorm.DB, filter domain.Filter) *gorm.DB {
	if filter.AllowAllBranches {
		return query
	}

	if len(filter.BranchIDs) == 0 {
		return query.Where("1 = 0")
	}

	return query.Where("branch_id IN ?", filter.BranchIDs)
}

func applyCustomerBranchFilter(query *gorm.DB, filter domain.Filter, column string) *gorm.DB {
	if filter.AllowAllBranches {
		return query
	}

	if len(filter.BranchIDs) == 0 {
		return query.Where("1 = 0")
	}

	return query.Where(column+" IN ?", filter.BranchIDs)
}
