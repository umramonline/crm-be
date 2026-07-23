package application

type customerEventPayload struct {
	UOId             uint64             `json:"uo_id"`
	BranchID         int32              `json:"branch_id"`
	Unvan            string             `json:"unvan"`
	Ad               string             `json:"ad"`
	Soyad            string             `json:"soyad"`
	YetkiliAdi       string             `json:"yetkili_adi"`
	Cep              string             `json:"cep"`
	Telefon          string             `json:"telefon"`
	Fax              string             `json:"fax"`
	Eposta           string             `json:"eposta"`
	Web              string             `json:"web"`
	Mahalle          string             `json:"mahalle"`
	Cadde            string             `json:"cadde"`
	Sokak            string             `json:"sokak"`
	Semt             string             `json:"semt"`
	KapiNo           string             `json:"kapi_no"`
	IlKodu           string             `json:"il_kodu"`
	IlceKodu         string             `json:"ilce_kodu"`
	Ulke             string             `json:"ulke"`
	DogumTarihi      *string            `json:"dogum_tarihi"`
	VadeGunu         *string            `json:"vade_gunu"`
	VergiDairesi     string             `json:"vergi_dairesi"`
	VergiDairesiKodu string             `json:"vergi_dairesi_kodu"`
	VergiNo          string             `json:"vergi_no"`
	TCNo             string             `json:"tc_no"`
	Type             string             `json:"type"`
	Mersis           string             `json:"mersis"`
	PasaportNo       string             `json:"pasaport_no"`
	PasaportBelge    string             `json:"pasaport_belge"`
	EsbisNo          string             `json:"esbis_no"`
	YetkiBelgeNo     string             `json:"yetki_belge_no"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
	Telephones       []telephonePayload `json:"telephones"`
	OccurredAt       string             `json:"occurred_at"`
}

type telephonePayload struct {
	PhoneNumber string `json:"phone_number"`
	Title       string `json:"title"`
}

type customerDeletedPayload struct {
	UOId       uint64 `json:"uo_id"`
	OccurredAt string `json:"occurred_at"`
}
