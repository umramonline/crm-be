package persistence

import (
	"context"
	"time"

	authapp "github.com/umran/new.crm/backend/internal/authorization/application"
	"github.com/umran/new.crm/backend/internal/authorization/domain"
	"gorm.io/gorm"
)

type ModuleModel struct {
	ID        uint64              `gorm:"primaryKey;autoIncrement"`
	Name      string              `gorm:"size:120;not null;uniqueIndex:idx_modules_name"`
	CreatedAt time.Time           `gorm:"precision:3"`
	UpdatedAt time.Time           `gorm:"precision:3"`
	DeletedAt gorm.DeletedAt      `gorm:"precision:3;index"`
	Methods   []ModuleMethodModel `gorm:"foreignKey:ModuleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ModuleModel) TableName() string {
	return "modules"
}

type ModuleMethodModel struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"`
	ModuleID    uint64         `gorm:"not null;index:idx_module_methods_module_id;uniqueIndex:idx_module_method_name;uniqueIndex:idx_module_method_path"`
	Name        string         `gorm:"size:120;not null;uniqueIndex:idx_module_method_name"`
	Description string         `gorm:"size:250"`
	Method      string         `gorm:"type:enum('GET','POST','PUT','PATCH','DELETE','HEAD','OPTIONS');default:null;uniqueIndex:idx_module_method_path"`
	Path        string         `gorm:"size:255;default:null;uniqueIndex:idx_module_method_path"`
	CreatedAt   time.Time      `gorm:"precision:3"`
	UpdatedAt   time.Time      `gorm:"precision:3"`
	DeletedAt   gorm.DeletedAt `gorm:"precision:3;index"`
	Module      ModuleModel    `gorm:"foreignKey:ModuleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ModuleMethodModel) TableName() string {
	return "module_methods"
}

type RolePermissionModel struct {
	ID             uint64            `gorm:"primaryKey;autoIncrement"`
	RoleID         uint64            `gorm:"not null;uniqueIndex:idx_role_module_method"`
	ModuleMethodID uint64            `gorm:"not null;index:fk_role_permissions_module_method;uniqueIndex:idx_role_module_method"`
	CreatedAt      time.Time         `gorm:"precision:3"`
	UpdatedAt      time.Time         `gorm:"precision:3"`
	DeletedAt      gorm.DeletedAt    `gorm:"precision:3;index"`
	ModuleMethod   ModuleMethodModel `gorm:"foreignKey:ModuleMethodID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (RolePermissionModel) TableName() string {
	return "role_permissions"
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ModuleModel{}, &ModuleMethodModel{}, &RolePermissionModel{})
}

func (r *Repository) ListPermissionsByRoleID(ctx context.Context, roleID uint64) ([]authapp.Permission, error) {
	var permissions []RolePermissionModel
	if err := r.db.WithContext(ctx).
		Preload("ModuleMethod.Module").
		Where("role_id = ?", roleID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	result := make([]authapp.Permission, 0, len(permissions))
	for _, permission := range permissions {
		method := permission.ModuleMethod
		result = append(result, authapp.Permission{
			ModuleID:       method.ModuleID,
			ModuleName:     method.Module.Name,
			ModuleMethodID: method.ID,
			Name:           method.Name,
			Description:    method.Description,
			Method:         method.Method,
			Path:           method.Path,
		})
	}

	return result, nil
}

func (r *Repository) RoleHasAccess(ctx context.Context, roleID uint64, method string, path string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RolePermissionModel{}).
		Joins("JOIN module_methods ON module_methods.id = role_permissions.module_method_id AND module_methods.deleted_at IS NULL").
		Where("role_permissions.role_id = ?", roleID).
		Where("module_methods.method = ? AND module_methods.path = ?", method, path).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) ListModules(ctx context.Context) ([]domain.Module, error) {
	var modules []ModuleModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&modules).Error; err != nil {
		return nil, err
	}

	result := make([]domain.Module, 0, len(modules))
	for _, module := range modules {
		result = append(result, moduleToDomain(module))
	}

	return result, nil
}

func (r *Repository) CreateModule(ctx context.Context, name string) (domain.Module, error) {
	module := ModuleModel{Name: name}
	if err := r.db.WithContext(ctx).Create(&module).Error; err != nil {
		return domain.Module{}, err
	}

	return moduleToDomain(module), nil
}

func (r *Repository) UpdateModule(ctx context.Context, id uint64, name string) (domain.Module, error) {
	var module ModuleModel
	if err := r.db.WithContext(ctx).First(&module, id).Error; err != nil {
		return domain.Module{}, err
	}

	module.Name = name
	if err := r.db.WithContext(ctx).Save(&module).Error; err != nil {
		return domain.Module{}, err
	}

	return moduleToDomain(module), nil
}

func (r *Repository) DeleteModule(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&ModuleModel{}, id).Error
}

func (r *Repository) ListModuleMethods(ctx context.Context, moduleID uint64) ([]domain.ModuleMethod, error) {
	query := r.db.WithContext(ctx).Preload("Module").Order("name ASC")
	if moduleID > 0 {
		query = query.Where("module_id = ?", moduleID)
	}

	var methods []ModuleMethodModel
	if err := query.Find(&methods).Error; err != nil {
		return nil, err
	}

	result := make([]domain.ModuleMethod, 0, len(methods))
	for _, method := range methods {
		result = append(result, moduleMethodToDomain(method))
	}

	return result, nil
}

func (r *Repository) CreateModuleMethod(ctx context.Context, input authapp.ModuleMethodInput) (domain.ModuleMethod, error) {
	method := ModuleMethodModel{
		ModuleID:    input.ModuleID,
		Name:        input.Name,
		Description: input.Description,
		Method:      input.Method,
		Path:        input.Path,
	}
	if err := r.db.WithContext(ctx).Create(&method).Error; err != nil {
		return domain.ModuleMethod{}, err
	}

	if err := r.db.WithContext(ctx).Preload("Module").First(&method, method.ID).Error; err != nil {
		return domain.ModuleMethod{}, err
	}

	return moduleMethodToDomain(method), nil
}

func (r *Repository) UpdateModuleMethod(ctx context.Context, id uint64, input authapp.ModuleMethodInput) (domain.ModuleMethod, error) {
	var method ModuleMethodModel
	if err := r.db.WithContext(ctx).First(&method, id).Error; err != nil {
		return domain.ModuleMethod{}, err
	}

	method.ModuleID = input.ModuleID
	method.Name = input.Name
	method.Description = input.Description
	method.Method = input.Method
	method.Path = input.Path
	if err := r.db.WithContext(ctx).Save(&method).Error; err != nil {
		return domain.ModuleMethod{}, err
	}

	if err := r.db.WithContext(ctx).Preload("Module").First(&method, method.ID).Error; err != nil {
		return domain.ModuleMethod{}, err
	}

	return moduleMethodToDomain(method), nil
}

func (r *Repository) DeleteModuleMethod(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&ModuleMethodModel{}, id).Error
}

func (r *Repository) ListRolePermissions(ctx context.Context, roleID uint64) ([]domain.RolePermission, error) {
	var permissions []RolePermissionModel
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&permissions).Error; err != nil {
		return nil, err
	}

	result := make([]domain.RolePermission, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, domain.RolePermission{
			ID:             permission.ID,
			RoleID:         permission.RoleID,
			ModuleMethodID: permission.ModuleMethodID,
			CreatedAt:      permission.CreatedAt,
			UpdatedAt:      permission.UpdatedAt,
		})
	}

	return result, nil
}

func (r *Repository) ReplaceRolePermissions(ctx context.Context, roleID uint64, moduleMethodIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermissionModel{}).Error; err != nil {
			return err
		}

		for _, moduleMethodID := range moduleMethodIDs {
			if moduleMethodID == 0 {
				continue
			}

			var permission RolePermissionModel
			err := tx.Unscoped().
				Where("role_id = ? AND module_method_id = ?", roleID, moduleMethodID).
				First(&permission).Error
			if err == nil {
				if err := tx.Unscoped().
					Model(&permission).
					Updates(map[string]any{"deleted_at": nil}).Error; err != nil {
					return err
				}

				continue
			}

			if err != gorm.ErrRecordNotFound {
				return err
			}

			permission = RolePermissionModel{
				RoleID:         roleID,
				ModuleMethodID: moduleMethodID,
			}
			if err := tx.Create(&permission).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func moduleToDomain(module ModuleModel) domain.Module {
	return domain.Module{
		ID:        module.ID,
		Name:      module.Name,
		CreatedAt: module.CreatedAt,
		UpdatedAt: module.UpdatedAt,
	}
}

func moduleMethodToDomain(method ModuleMethodModel) domain.ModuleMethod {
	return domain.ModuleMethod{
		ID:          method.ID,
		ModuleID:    method.ModuleID,
		ModuleName:  method.Module.Name,
		Name:        method.Name,
		Description: method.Description,
		Method:      method.Method,
		Path:        method.Path,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	}
}
