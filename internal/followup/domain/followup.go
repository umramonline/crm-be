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
	VisitType              string       `json:"visit_type"`
	VisitDate              string       `json:"visit_date"`
	NextVisitDate          string       `json:"next_visit_date"`
	AgreementReached       bool         `json:"agreement_reached"`
	AgreementFailureReason string       `json:"agreement_failure_reason,omitempty"`
	Note                   string       `json:"note,omitempty"`
	Images                 []Image      `json:"images"`
	MeetPeople             []MeetPerson `json:"meet_people"`
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
