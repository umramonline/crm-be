package persistence

import "gorm.io/gorm"

var followUpMethodSeeds = []authorizationMethodSeed{
	{
		Name:        "follow_ups.menu",
		Description: "Sol menüde Tüm Takip Kayıtları menüsünü gösterir.",
	},
	{
		Name:        "follow_ups.list",
		Description: "Takip kayıtları listesini görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/follow-ups"),
	},
	{
		Name:        "follow_ups.assigned.list",
		Description: "Kullanıcının kendisine atanmış takip kayıtlarını görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/follow-ups/assigned-to-me"),
	},
	{
		Name:        "follow_ups.detail",
		Description: "Takip kaydı detayını görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/follow-ups/:uuid"),
	},
	{
		Name:        "follow_ups.create",
		Description: "Görev müşterisi için takip kaydı oluşturur.",
		Method:      stringPointer("POST"),
		Path:        stringPointer("/api/v1/follow-ups"),
	},
}

func SeedFollowUps(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		module, err := seedFollowUpsModule(tx)
		if err != nil {
			return err
		}

		for _, methodSeed := range followUpMethodSeeds {
			method, err := seedFollowUpAuthorizationMethod(tx, module.ID, methodSeed)
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

func seedFollowUpAuthorizationMethod(tx *gorm.DB, moduleID uint64, seed authorizationMethodSeed) (ModuleMethodModel, error) {
	var method ModuleMethodModel
	err := tx.Unscoped().Where("name = ?", seed.Name).First(&method).Error
	if err == gorm.ErrRecordNotFound && seed.Method != nil && seed.Path != nil {
		err = tx.Unscoped().
			Where("method = ? AND path = ?", *seed.Method, *seed.Path).
			First(&method).Error
	}

	if err == nil {
		updates := map[string]any{
			"module_id":   moduleID,
			"name":        seed.Name,
			"description": seed.Description,
			"method":      seed.Method,
			"path":        seed.Path,
			"deleted_at":  nil,
		}
		if err := tx.Unscoped().Model(&method).Updates(updates).Error; err != nil {
			return ModuleMethodModel{}, err
		}

		method.ModuleID = moduleID
		method.Name = seed.Name
		method.Description = seed.Description
		method.Method = seed.Method
		method.Path = seed.Path
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

func seedFollowUpsModule(tx *gorm.DB) (ModuleModel, error) {
	var module ModuleModel
	err := tx.Unscoped().Where("name = ?", "follow_ups").First(&module).Error
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

	module = ModuleModel{Name: "follow_ups"}
	if err := tx.Create(&module).Error; err != nil {
		return ModuleModel{}, err
	}

	return module, nil
}
