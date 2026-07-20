package domain

import "time"

type Filter struct {
	StartDate        time.Time
	EndDate          time.Time
	BranchIDs        []uint64
	AllowAllBranches bool
}

type TaskStats struct {
	PendingCount    int64
	InProgressCount int64
	CompletedCount  int64
}

type Stats struct {
	PotentialCustomerCount int64   `json:"potential_customer_count"`
	TotalCustomerCount     int64   `json:"total_customer_count"`
	CustomerVisitCount     int64   `json:"customer_visit_count"`
	NewCustomerCount       int64   `json:"new_customer_count"`
	VehicleEntryCount      int64   `json:"vehicle_entry_count"`
	TotalAmount            float64 `json:"total_amount"`
	LoadedCreditAmount     float64 `json:"loaded_credit_amount"`
	VehicleStockCount      int64   `json:"vehicle_stock_count"`
	PendingTaskCount       int64   `json:"pending_task_count"`
	InProgressTaskCount    int64   `json:"in_progress_task_count"`
	CompletedTaskCount     int64   `json:"completed_task_count"`
	OverdueTaskCount       int64   `json:"overdue_task_count"`
}
