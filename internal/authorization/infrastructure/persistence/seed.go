package persistence

import "gorm.io/gorm"

const adminRoleID uint64 = 30

type authorizationMethodSeed struct {
	Name        string
	Description string
	Method      string
	Path        string
}

var authorizationMethodSeeds = []authorizationMethodSeed{
	{
		Name:        "Rol Listesi",
		Description: "Umramonline aktif rollerini listeler.",
		Method:      "GET",
		Path:        "/api/v1/authorization/roles",
	},
	{
		Name:        "Modül Listesi",
		Description: "Authorization modüllerini listeler.",
		Method:      "GET",
		Path:        "/api/v1/authorization/modules",
	},
	{
		Name:        "Modül Oluşturma",
		Description: "Authorization modülü oluşturur.",
		Method:      "POST",
		Path:        "/api/v1/authorization/modules",
	},
	{
		Name:        "Modül Güncelleme",
		Description: "Authorization modülünü günceller.",
		Method:      "PUT",
		Path:        "/api/v1/authorization/modules/:id",
	},
	{
		Name:        "Modül Silme",
		Description: "Authorization modülünü siler.",
		Method:      "DELETE",
		Path:        "/api/v1/authorization/modules/:id",
	},
	{
		Name:        "Modül Method Listesi",
		Description: "Authorization modül methodlarını listeler.",
		Method:      "GET",
		Path:        "/api/v1/authorization/module-methods",
	},
	{
		Name:        "Modül Method Oluşturma",
		Description: "Authorization modül methodu oluşturur.",
		Method:      "POST",
		Path:        "/api/v1/authorization/module-methods",
	},
	{
		Name:        "Modül Method Güncelleme",
		Description: "Authorization modül methodunu günceller.",
		Method:      "PUT",
		Path:        "/api/v1/authorization/module-methods/:id",
	},
	{
		Name:        "Modül Method Silme",
		Description: "Authorization modül methodunu siler.",
		Method:      "DELETE",
		Path:        "/api/v1/authorization/module-methods/:id",
	},
	{
		Name:        "Rol İzin Listesi",
		Description: "Role ait authorization izinlerini listeler.",
		Method:      "GET",
		Path:        "/api/v1/authorization/role-permissions",
	},
	{
		Name:        "Rol İzin Güncelleme",
		Description: "Role ait authorization izinlerini topluca günceller.",
		Method:      "PUT",
		Path:        "/api/v1/authorization/role-permissions/:role_id",
	},
}

func SeedAuthorization(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		module, err := seedAuthorizationModule(tx)
		if err != nil {
			return err
		}

		for _, methodSeed := range authorizationMethodSeeds {
			method, err := seedAuthorizationMethod(tx, module.ID, methodSeed)
			if err != nil {
				return err
			}

			if err := seedAdminPermission(tx, method.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

func seedAuthorizationModule(tx *gorm.DB) (ModuleModel, error) {
	var module ModuleModel
	err := tx.Unscoped().Where("name = ?", "authorization").First(&module).Error
	if err == nil {
		if module.DeletedAt.Valid {
			if err := tx.Unscoped().Model(&module).Update("deleted_at", nil).Error; err != nil {
				return ModuleModel{}, err
			}
			module.DeletedAt = gorm.DeletedAt{}
		}

		return module, nil
	}

	if err != gorm.ErrRecordNotFound {
		return ModuleModel{}, err
	}

	module = ModuleModel{Name: "authorization"}
	if err := tx.Create(&module).Error; err != nil {
		return ModuleModel{}, err
	}

	return module, nil
}

func seedAuthorizationMethod(tx *gorm.DB, moduleID uint64, seed authorizationMethodSeed) (ModuleMethodModel, error) {
	var method ModuleMethodModel
	err := tx.Unscoped().
		Where("module_id = ? AND method = ? AND path = ?", moduleID, seed.Method, seed.Path).
		First(&method).Error
	if err == nil {
		updates := map[string]any{
			"name":        seed.Name,
			"description": seed.Description,
			"deleted_at":  nil,
		}
		if err := tx.Unscoped().Model(&method).Updates(updates).Error; err != nil {
			return ModuleMethodModel{}, err
		}

		method.Name = seed.Name
		method.Description = seed.Description
		method.DeletedAt = gorm.DeletedAt{}

		return method, nil
	}

	if err != gorm.ErrRecordNotFound {
		return ModuleMethodModel{}, err
	}

	method = ModuleMethodModel{
		ModuleID:    moduleID,
		Name:        seed.Name,
		Description: seed.Description,
		Method:      seed.Method,
		Path:        seed.Path,
	}
	if err := tx.Create(&method).Error; err != nil {
		return ModuleMethodModel{}, err
	}

	return method, nil
}

func seedAdminPermission(tx *gorm.DB, moduleMethodID uint64) error {
	var permission RolePermissionModel
	err := tx.Unscoped().
		Where("role_id = ? AND module_method_id = ?", adminRoleID, moduleMethodID).
		First(&permission).Error
	if err == nil {
		if permission.DeletedAt.Valid {
			return tx.Unscoped().Model(&permission).Update("deleted_at", nil).Error
		}

		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	return tx.Create(&RolePermissionModel{
		RoleID:         adminRoleID,
		ModuleMethodID: moduleMethodID,
	}).Error
}
