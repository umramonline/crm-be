package domain

type Customer struct {
	ID           uint64  `json:"id"`
	Situation    string  `json:"situation"`
	Unvan        string  `json:"unvan"`
	Cep          string  `json:"cep"`
	Ad           string  `json:"ad"`
	Soyad        string  `json:"soyad"`
	BranchName   string  `json:"branch_name"`
	ZoneName     string  `json:"zone_name"`
	PlusCardNo   string  `json:"plus_card_no"`
	Credit       string  `json:"credit"`
	Source       string  `json:"source"`
	City         string  `json:"city"`
	Town         string  `json:"town"`
	CreatedAt    *string `json:"created_at,omitempty"`
	Type         string  `json:"type"`
	DaysSpending *int    `json:"days_spending,omitempty"`
	DaysLoading  *int    `json:"days_loading,omitempty"`
}

type CustomerDetail struct {
	ID                     uint64              `json:"id"`
	UOId                   uint64              `json:"uo_id,omitempty"`
	BranchID               *int32              `json:"branch_id,omitempty"`
	Unvan                  string              `json:"unvan"`
	Ad                     string              `json:"ad"`
	Soyad                  string              `json:"soyad"`
	YetkiliAdi             string              `json:"yetkili_adi"`
	Cep                    string              `json:"cep"`
	Telefon                string              `json:"telefon"`
	Eposta                 string              `json:"eposta"`
	Website                string              `json:"website"`
	GoogleMapLink          string              `json:"google_map_link"`
	ClassifiedsWebsiteLink string              `json:"classifieds_website_link"`
	Mahalle                string              `json:"mahalle"`
	AddressDetail          string              `json:"address_detail"`
	IlKodu                 string              `json:"il_kodu"`
	IlceKodu               string              `json:"ilce_kodu"`
	VergiNo                string              `json:"vergi_no"`
	TCNo                   string              `json:"tc_no"`
	DogumTarihi            string              `json:"dogum_tarihi"`
	VehicleStockCount      *int32              `json:"vehicle_stock_count,omitempty"`
	Type                   string              `json:"type"`
	CreatedAt              *string             `json:"created_at,omitempty"`
	Telephones             []CustomerTelephone `json:"telephones,omitempty"`
}

type CustomerSearchResult struct {
	Found    bool            `json:"found"`
	Source   string          `json:"source,omitempty"`
	Customer *CustomerDetail `json:"customer,omitempty"`
}

type CreateCustomerInput struct {
	Type       string
	Ad         string
	Soyad      string
	Cep        string
	Unvan      string
	YetkiliAdi string
	Telefon    string
	IlKodu     string
	IlceKodu   string
	Mahalle    string
	BranchID   int32
}

type CustomerTelephone struct {
	ID          uint64 `json:"id,omitempty"`
	PhoneNumber string `json:"phone_number"`
	Title       string `json:"title"`
}

type FullRegistrationInput struct {
	Type                   string
	Cep                    string
	Ad                     string
	Soyad                  string
	TCNo                   string
	DogumTarihi            string
	Eposta                 string
	Website                string
	GoogleMapLink          string
	ClassifiedsWebsiteLink string
	VehicleStockCount      int32
	BranchID               int32
	Telephones             []CustomerTelephone
	IlKodu                 string
	IlceKodu               string
	Mahalle                string
	AddressDetail          string
}

type City struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

type Town struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	CityID    uint64 `json:"city_id"`
	CityTitle string `json:"city_title,omitempty"`
}

type Branch struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

type Pagination struct {
	CurrentPage int  `json:"current_page"`
	LastPage    int  `json:"last_page"`
	PerPage     int  `json:"per_page"`
	Total       int  `json:"total"`
	From        *int `json:"from,omitempty"`
	To          *int `json:"to,omitempty"`
}

type ListResult struct {
	Items      []Customer `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type ListQuery struct {
	Page       int
	PerPage    int
	DataSource string
	Situation  string
	Unvan      string
	Cep        string
	Ad         string
	Soyad      string
	BranchName string
	PlusCardNo string
	Source     string
	City       string
	Town       string
	CreatedAt  string
	Type       string
	SortBy     string
	SortOrder  string
	ZoneID     int
	BranchIDs  []int32
	CityIDs    []string
	TownIDs    []string
}
