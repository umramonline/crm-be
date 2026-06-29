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
	{
		Name:        "customers.list.umramonline",
		Description: "Umramonline müşteri listesini görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/umramonline"),
	},
	{
		Name:        "customers.list.backend",
		Description: "Backend veritabanındaki müşteri listesini görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/backend"),
	},
	{
		Name:        "customers.search",
		Description: "Müşteri giriş ekranında müşteriyi arar.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/search"),
	},
	{
		Name:        "customers.detail",
		Description: "Müşteri detayını görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/:id"),
	},
	{
		Name:        "customers.detail.umramonline",
		Description: "Umramonline müşteri detayını görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/umramonline/:id"),
	},
	{
		Name:        "customers.detail.backend",
		Description: "Backend veritabanındaki müşteri detayını görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/backend/:id"),
	},
	{
		Name:        "customers.create",
		Description: "Müşteri giriş ekranından yeni müşteri oluşturur.",
		Method:      stringPointer("POST"),
		Path:        stringPointer("/api/v1/customers"),
	},
	{
		Name:        "customers.full_registration.detail",
		Description: "Müşteri tam kayıt bilgilerini görüntüler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/full-registration/:id"),
	},
	{
		Name:        "customers.full_registration.phone_exists",
		Description: "Müşteri tam kayıt cep telefonu benzersizliğini kontrol eder.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/customers/full-registration/:id/phone-exists"),
	},
	{
		Name:        "customers.full_registration.update",
		Description: "Müşteri tam kayıt bilgilerini günceller.",
		Method:      stringPointer("PUT"),
		Path:        stringPointer("/api/v1/customers/full-registration/:id"),
	},
	{
		Name:        "customers.zones.list",
		Description: "Müşteri filtresi için bölgeleri listeler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/zones"),
	},
	{
		Name:        "customers.cities.list",
		Description: "Müşteri giriş formu için şehirleri listeler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/cities"),
	},
	{
		Name:        "customers.towns.list",
		Description: "Müşteri giriş formu için ilçeleri listeler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/towns"),
	},
	{
		Name:        "customers.branches.list",
		Description: "Müşteri giriş formu için bayileri listeler.",
		Method:      stringPointer("GET"),
		Path:        stringPointer("/api/v1/branches"),
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
