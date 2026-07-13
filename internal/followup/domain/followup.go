package domain

type CreateFollowUpInput struct {
	AuthenticatedUserID    uint64
	TasksCustomerUUID      string
	VisitType              string
	VisitDate              string
	NextVisitDate          string
	AgreementReached       *bool
	AgreementFailureReason string
	Note                   string
	Images                 []ImageUpload
	MeetPeople             []MeetPersonInput
}

type ImageUpload struct {
	FileName    string
	ContentType string
	Size        int64
	Content     []byte
}

type StoredImage struct {
	UUID string
	Path string
	URL  string
}

type MeetPersonInput struct {
	Title   string
	Name    string
	Surname string
	Phone   string
	Email   string
}

type TaskCustomer struct {
	ID             uint64
	UUID           string
	Status         string
	AssignedUserID uint64
}

type PersistFollowUpInput struct {
	UUID                   string
	TasksCustomerID        uint64
	TasksCustomerUUID      string
	VisitType              string
	VisitDate              string
	NextVisitDate          string
	AgreementReached       bool
	AgreementFailureReason string
	Note                   string
	Images                 []StoredImage
	MeetPeople             []MeetPersonInput
}

type FollowUp struct {
	UUID                   string       `json:"uuid"`
	TasksCustomerUUID      string       `json:"tasks_customer_uuid"`
	TaskUUID               string       `json:"task_uuid,omitempty"`
	Title                  string       `json:"title,omitempty"`
	CustomerID             uint64       `json:"customer_id,omitempty"`
	CustomerUnvan          string       `json:"customer_unvan,omitempty"`
	AssignedUserFullName   string       `json:"assigned_user_full_name,omitempty"`
	BranchName             string       `json:"branch_name,omitempty"`
	VisitType              string       `json:"visit_type"`
	VisitDate              string       `json:"visit_date"`
	NextVisitDate          string       `json:"next_visit_date"`
	AgreementReached       bool         `json:"agreement_reached"`
	AgreementFailureReason string       `json:"agreement_failure_reason,omitempty"`
	Note                   string       `json:"note,omitempty"`
	Images                 []Image      `json:"images"`
	MeetPeople             []MeetPerson `json:"meet_people"`
}

type FollowUpListItem struct {
	UUID                 string `json:"uuid"`
	TasksCustomerUUID    string `json:"tasks_customer_uuid"`
	TaskUUID             string `json:"task_uuid"`
	Title                string `json:"title"`
	CustomerID           uint64 `json:"customer_id"`
	CustomerUnvan        string `json:"customer_unvan"`
	AssignedUserFullName string `json:"assigned_user_full_name"`
	BranchName           string `json:"branch_name"`
	VisitDate            string `json:"visit_date"`
	NextVisitDate        string `json:"next_visit_date"`
	AgreementReached     bool   `json:"agreement_reached"`
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
	Items      []FollowUpListItem `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

type ListQuery struct {
	Page                 int
	PerPage              int
	Title                string
	Customer             string
	AssignedUserFullName string
	BranchName           string
	VisitDate            string
	NextVisitDate        string
	SortBy               string
	SortOrder            string
}

type Image struct {
	UUID string `json:"uuid"`
	URL  string `json:"url"`
}

type MeetPerson struct {
	UUID    string `json:"uuid"`
	Title   string `json:"title"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Phone   string `json:"phone"`
	Email   string `json:"email,omitempty"`
}
