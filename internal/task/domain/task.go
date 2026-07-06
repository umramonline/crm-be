package domain

type Task struct {
	ID             uint64   `json:"id"`
	UUID           string   `json:"uuid"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	AssignedUserID uint64   `json:"assigned_user_id"`
	BranchID       uint64   `json:"branch_id"`
	BranchName     string   `json:"branch_name"`
	VisitDate      string   `json:"visit_date,omitempty"`
	DueDate        string   `json:"due_date,omitempty"`
	Status         string   `json:"status"`
	Priority       string   `json:"priority"`
	CustomerIDs    []uint64 `json:"customer_ids"`
}

type CreateTaskInput struct {
	Title                 string
	CreatedByUserID       uint64
	CreatedByUserFullName string
	Description           string
	AssignedUserID        uint64
	AssignedUserFullName  string
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
