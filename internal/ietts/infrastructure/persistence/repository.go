package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/umran/new.crm/backend/internal/ietts/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").
		AutoMigrate(&IettsRecordModel{})
}

func (r *Repository) ListRecords(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if r == nil || r.db == nil {
		return domain.ListResult{}, gorm.ErrInvalidDB
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	perPage := query.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	var total int64
	if err := r.listBaseQuery(ctx, query).Count(&total).Error; err != nil {
		return domain.ListResult{}, err
	}

	var rows []IettsRecordModel
	if err := r.listBaseQuery(ctx, query).
		Order(iettsListOrder(query)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&rows).Error; err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.RecordListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRecordListItem(row))
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage <= 0 {
		lastPage = 1
	}

	var from *int
	var to *int
	if total > 0 {
		fromValue := ((page - 1) * perPage) + 1
		toValue := fromValue + len(items) - 1
		from = &fromValue
		to = &toValue
	}

	return domain.ListResult{
		Items: items,
		Pagination: domain.Pagination{
			CurrentPage: page,
			LastPage:    lastPage,
			PerPage:     perPage,
			Total:       int(total),
			From:        from,
			To:          to,
		},
	}, nil
}

func (r *Repository) listBaseQuery(ctx context.Context, filters domain.ListQuery) *gorm.DB {
	query := r.db.WithContext(ctx).
		Model(&IettsRecordModel{}).
		Where("deleted_at IS NULL")

	return applyIettsFilters(query, filters)
}

func (r *Repository) FindRecordByUUID(ctx context.Context, uuid string) (domain.Record, error) {
	if r == nil || r.db == nil {
		return domain.Record{}, gorm.ErrInvalidDB
	}

	var row IettsRecordModel
	err := r.db.WithContext(ctx).
		Where("uuid = ? AND deleted_at IS NULL", uuid).
		First(&row).Error
	if err != nil {
		return domain.Record{}, err
	}

	return domain.Record{
		UUID:            domain.StringValue(row.UUID),
		CompanyName:     domain.StringValue(row.CompanyName),
		BusinessName:    domain.StringValue(row.BusinessName),
		BusinessAddress: domain.StringValue(row.BusinessAddress),
		CustomerID:      customerIDValue(row.CustomerID),
	}, nil
}

func (r *Repository) UpdateCustomerIDByUUID(ctx context.Context, uuid string, customerID uint64) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}

	result := r.db.WithContext(ctx).
		Model(&IettsRecordModel{}).
		Where("uuid = ? AND deleted_at IS NULL", uuid).
		Update("customer_id", customerID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func customerIDValue(value *uint64) uint64 {
	if value == nil {
		return 0
	}

	return *value
}

func applyIettsFilters(query *gorm.DB, filters domain.ListQuery) *gorm.DB {
	if strings.TrimSpace(filters.DocumentNumber) != "" {
		query = query.Where("document_number LIKE ?", "%"+strings.TrimSpace(filters.DocumentNumber)+"%")
	}

	if strings.TrimSpace(filters.CompanyName) != "" {
		query = query.Where("company_name LIKE ?", "%"+strings.TrimSpace(filters.CompanyName)+"%")
	}

	if strings.TrimSpace(filters.BusinessName) != "" {
		query = query.Where("business_name LIKE ?", "%"+strings.TrimSpace(filters.BusinessName)+"%")
	}

	if strings.TrimSpace(filters.BusinessAddress) != "" {
		query = query.Where("business_address LIKE ?", "%"+strings.TrimSpace(filters.BusinessAddress)+"%")
	}

	if strings.TrimSpace(filters.DocumentIssueDate) != "" {
		query = query.Where("document_issue_date LIKE ?", "%"+strings.TrimSpace(filters.DocumentIssueDate)+"%")
	}

	if strings.TrimSpace(filters.DocumentStatus) != "" {
		query = query.Where("document_status LIKE ?", "%"+strings.TrimSpace(filters.DocumentStatus)+"%")
	}

	if strings.TrimSpace(filters.City) != "" {
		query = query.Where("city LIKE ?", "%"+strings.TrimSpace(filters.City)+"%")
	}

	if strings.TrimSpace(filters.District) != "" {
		query = query.Where("district LIKE ?", "%"+strings.TrimSpace(filters.District)+"%")
	}

	if strings.TrimSpace(filters.CreatedAt) != "" {
		query = query.Where("created_at LIKE ?", "%"+strings.TrimSpace(filters.CreatedAt)+"%")
	}

	return query
}

func iettsListOrder(query domain.ListQuery) string {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	switch sortBy {
	case "document_issue_date", "created_at":
	default:
		return "id DESC"
	}

	sortOrder := "DESC"
	if strings.ToLower(strings.TrimSpace(query.SortOrder)) == "asc" {
		sortOrder = "ASC"
	}

	return sortBy + " " + sortOrder + ", id DESC"
}

func toRecordListItem(model IettsRecordModel) domain.RecordListItem {
	item := domain.RecordListItem{
		DocumentNumber:  model.DocumentNumber,
		CompanyName:     model.CompanyName,
		BusinessName:    model.BusinessName,
		BusinessAddress: model.BusinessAddress,
		DocumentStatus:  model.DocumentStatus,
		City:            model.City,
		District:        model.District,
	}

	if model.UUID != nil {
		item.UUID = strings.TrimSpace(*model.UUID)
	}

	if model.DocumentIssueDate != nil {
		item.DocumentIssueDate = model.DocumentIssueDate.Format(time.DateOnly)
	}

	if model.CreatedAt != nil {
		item.CreatedAt = model.CreatedAt.Format(time.RFC3339)
	}

	if model.CustomerID != nil && *model.CustomerID > 0 {
		item.CustomerID = model.CustomerID
	}

	return item
}
