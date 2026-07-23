package domain

const (
	EventTypeCustomerCreated = "customer.created"
	EventTypeCustomerUpdated = "customer.updated"
)

type ConsumeCommand struct {
	EventID   string
	EventType string
	Payload   []byte
}

type Telephone struct {
	PhoneNumber string
	Title       string
}

type CustomerEvent struct {
	EventID          string
	EventType        string
	UOId             uint64
	BranchID         int32
	Unvan            string
	Ad               string
	Soyad            string
	YetkiliAdi       string
	Cep              string
	Telefon          string
	Fax              string
	Eposta           string
	Web              string
	Mahalle          string
	Cadde            string
	Sokak            string
	Semt             string
	KapiNo           string
	IlKodu           string
	IlceKodu         string
	Ulke             string
	DogumTarihi      *string
	VadeGunu         *string
	VergiDairesi     string
	VergiDairesiKodu string
	VergiNo          string
	TCNo             string
	Type             string
	Mersis           string
	PasaportNo       string
	PasaportBelge    string
	EsbisNo          string
	YetkiBelgeNo     string
	CreatedAt        string
	UpdatedAt        string
	Telephones       []Telephone
	OccurredAt       string
}

type CustomerCreatedEvent = CustomerEvent

type CustomerUpdatedEvent = CustomerEvent

type ConsumeResult struct {
	EventID    string
	CustomerID uint64
	Action     string
}
