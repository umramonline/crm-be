package domain

import "time"

type Module struct {
	ID        uint64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModuleMethod struct {
	ID          uint64
	ModuleID    uint64
	ModuleName  string
	Name        string
	Description string
	Method      string
	Path        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RolePermission struct {
	ID             uint64
	RoleID         uint64
	ModuleMethodID uint64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Role struct {
	ID   uint64
	Name string
}

type User struct {
	ID       uint64
	Name     string
	Phone    string
	RoleID   uint64
	RoleName string
}
