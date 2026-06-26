package domain

type Customer struct {
	Situation    string  `json:"situation"`
	Unvan        string  `json:"unvan"`
	Cep          string  `json:"cep"`
	Ad           string  `json:"ad"`
	Soyad        string  `json:"soyad"`
	BranchName   string  `json:"branch_name"`
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
}
