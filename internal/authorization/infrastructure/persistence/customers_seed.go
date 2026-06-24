package persistence

import "gorm.io/gorm"

var customerMethodSeeds = []authorizationMethodSeed{
	{
		Name:        "customers.menu",
		Description: "Sol menüde Müşteriler menüsünü gösterir.",
	},
	{
		Name:        "customers.list",
		Description: "Müşteri listesini görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers"),
	},
}

func SeedCustomers(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		module, err := seedCustomersModule(tx)
		if err != nil {
			return err
		}

		for _, methodSeed := range customerMethodSeeds {
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

func seedCustomersModule(tx *gorm.DB) (ModuleModel, error) {
	var module ModuleModel
	err := tx.Unscoped().Where("name = ?", "customers").First(&module).Error
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

	module = ModuleModel{Name: "customers"}
	if err := tx.Create(&module).Error; err != nil {
		return ModuleModel{}, err
	}

	return module, nil
}
