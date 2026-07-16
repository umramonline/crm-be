package domain

type RecordListItem struct {
	UUID              string  `json:"uuid"`
	DocumentNumber    string  `json:"document_number"`
	CompanyName       *string `json:"company_name,omitempty"`
	BusinessName      *string `json:"business_name,omitempty"`
	BusinessAddress   *string `json:"business_address,omitempty"`
	DocumentIssueDate string  `json:"document_issue_date,omitempty"`
	DocumentStatus    *string `json:"document_status,omitempty"`
	City              *string `json:"city,omitempty"`
	District          *string `json:"district,omitempty"`
	CreatedAt         string  `json:"created_at,omitempty"`
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
	Items      []RecordListItem `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type ListQuery struct {
	Page              int
	PerPage           int
	DocumentNumber    string
	CompanyName       string
	BusinessName      string
	BusinessAddress   string
	DocumentIssueDate string
	DocumentStatus    string
	City              string
	District          string
	CreatedAt         string
	SortBy            string
	SortOrder         string
}
