package persistence

import "gorm.io/gorm"

var iettsMethodSeeds = []authorizationMethodSeed{
	{
		Name:        "ietts.menu",
		Description: "Sol menüde IETTS menüsünü gösterir.",
	},
	{
		Name:        "ietts.list",
		Description: "IETTS kayıtlarını listeler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/ietts"),
	},
	{
		Name:        "ietts.convert_to_customer",
		Description: "IETTS kaydını müşteriye dönüştürür.",
		Method:      stringPointer("POST"),
		Path:        stringPointer("/api/v1/ietts/:uuid/convert-to-customer"),
	},
}

func SeedIetts(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		module, err := seedIettsModule(tx)
		if err != nil {
			return err
		}

		for _, methodSeed := range iettsMethodSeeds {
			method, err := seedIettsAuthorizationMethod(tx, module.ID, methodSeed)
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

func seedIettsAuthorizationMethod(tx *gorm.DB, moduleID uint64, seed authorizationMethodSeed) (ModuleMethodModel, error) {
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

func seedIettsModule(tx *gorm.DB) (ModuleModel, error) {
	var module ModuleModel
	err := tx.Unscoped().Where("name = ?", "ietts").First(&module).Error
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

	module = ModuleModel{Name: "ietts"}
	if err := tx.Create(&module).Error; err != nil {
		return ModuleModel{}, err
	}

	return module, nil
}
