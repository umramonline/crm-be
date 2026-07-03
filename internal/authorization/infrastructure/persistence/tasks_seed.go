package persistence

import "gorm.io/gorm"

var taskMethodSeeds = []authorizationMethodSeed{
	{
		Name:        "tasks.create",
		Description: "Seçili müşteriler için görev oluşturur.",
		Method:      stringPointer("POST"),
		Path:        stringPointer("/api/v1/tasks"),
	},
}

func SeedTasks(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		module, err := seedTasksModule(tx)
		if err != nil {
			return err
		}

		for _, methodSeed := range taskMethodSeeds {
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

func seedTasksModule(tx *gorm.DB) (ModuleModel, error) {
	var module ModuleModel
	err := tx.Unscoped().Where("name = ?", "tasks").First(&module).Error
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

	module = ModuleModel{Name: "tasks"}
	if err := tx.Create(&module).Error; err != nil {
		return ModuleModel{}, err
	}

	return module, nil
}
