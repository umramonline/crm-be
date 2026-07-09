package domain

type Task struct {
	ID                    uint64   `json:"-"`
	UUID                  string   `json:"uuid"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	CreatedByUserID       uint64   `json:"created_by_user_id"`
	CreatedByUserFullName string   `json:"created_by_user_full_name"`
	AssignedUserID        uint64   `json:"assigned_user_id"`
	AssignedUserFullName  string   `json:"assigned_user_full_name"`
	AssignedUserPhone     string   `json:"-"`
	BranchID              uint64   `json:"branch_id"`
	BranchName            string   `json:"branch_name"`
	VisitDate             string   `json:"visit_date,omitempty"`
	DueDate               string   `json:"due_date,omitempty"`
	Priority              string   `json:"priority"`
	CustomerIDs           []uint64 `json:"customer_ids"`
}

type CreateTaskInput struct {
	Title                 string
	CreatedByUserID       uint64
	CreatedByUserFullName string
	Description           string
	AssignedUserID        uint64
	AssignedUserFullName  string
	AssignedUserPhone     string
	BranchID              uint64
	BranchName            string
	VisitDate             string
	DueDate               string
	Priority              string
	CustomerIDs           []uint64
}

type Branch struct {
	ID    uint64
	Name  string
	Title string
}

type BranchUser struct {
	ID    uint64
	Name  string
	Phone string
}

type TaskCreatedSMSInput struct {
	Phone                string
	TaskUUID             string
	Title                string
	AssignedUserFullName string
	BranchName           string
	VisitDate            string
	DueDate              string
	Priority             string
}

type TaskCustomer struct {
	ID         uint64 `json:"-"`
	UUID       string `json:"uuid"`
	CustomerID uint64 `json:"customer_id"`
	Unvan      string `json:"unvan"`
	Ad         string `json:"ad"`
	Soyad      string `json:"soyad"`
	Status     string `json:"status"`
}

type TaskListItem struct {
	UUID                  string         `json:"uuid"`
	Title                 string         `json:"title"`
	Description           string         `json:"description"`
	CreatedByUserFullName string         `json:"created_by_user_full_name"`
	AssignedUserFullName  string         `json:"assigned_user_full_name"`
	AssignedUserID        uint64         `json:"assigned_user_id"`
	BranchName            string         `json:"branch_name"`
	VisitDate             string         `json:"visit_date,omitempty"`
	DueDate               string         `json:"due_date,omitempty"`
	Status                string         `json:"status,omitempty"`
	Priority              string         `json:"priority"`
	CustomerCount         int            `json:"customer_count"`
	Customers             []TaskCustomer `json:"customers"`
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
	Items      []TaskListItem `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type ListQuery struct {
	Page                  int
	PerPage               int
	Title                 string
	Customer              string
	AssignedUserID        uint64
	AssignedUserFullName  string
	BranchName            string
	VisitDate             string
	DueDate               string
	Priority              string
	CreatedByUserFullName string
	SortBy                string
	SortOrder             string
}
