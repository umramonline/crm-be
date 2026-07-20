package application

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/umran/new.crm/backend/internal/authorization/domain"
	sharedauth "github.com/umran/new.crm/backend/internal/shared/auth"
)

var (
	ErrAuthorizationUnavailable = errors.New("authorization unavailable")
	ErrInvalidRole              = errors.New("invalid role")
	ErrInvalidUser              = errors.New("invalid user")
)

type Permission struct {
	ModuleID       uint64 `json:"module_id"`
	ModuleName     string `json:"module_name"`
	ModuleMethodID uint64 `json:"module_method_id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Method         string `json:"method,omitempty"`
	Path           string `json:"path,omitempty"`
}

type User struct {
	ID        uint64   `json:"id"`
	FullName  string   `json:"full_name,omitempty"`
	Phone     string   `json:"phone,omitempty"`
	RoleID    uint64   `json:"role_id"`
	RoleName  string   `json:"role_name,omitempty"`
	BranchIds []uint64 `json:"branch_ids,omitempty"`
	Branches  []Branch `json:"branches,omitempty"`
}

type Branch struct {
	ID     uint64 `json:"id"`
	KisaAd string `json:"kisa_ad"`
}

type SessionData struct {
	UserID      string       `json:"user_id"`
	User        User         `json:"user"`
	Permissions []Permission `json:"permissions"`
}

type RoleProvider interface {
	ListRoles(ctx context.Context) ([]domain.Role, error)
}

type PermissionRepository interface {
	ListPermissionsByRoleID(ctx context.Context, roleID uint64) ([]Permission, error)
	RoleHasAccess(ctx context.Context, roleID uint64, method string, path string) (bool, error)
}

type ModuleRepository interface {
	ListModules(ctx context.Context) ([]domain.Module, error)
	CreateModule(ctx context.Context, name string) (domain.Module, error)
	UpdateModule(ctx context.Context, id uint64, name string) (domain.Module, error)
	DeleteModule(ctx context.Context, id uint64) error
	ListModuleMethods(ctx context.Context, moduleID uint64) ([]domain.ModuleMethod, error)
	CreateModuleMethod(ctx context.Context, input ModuleMethodInput) (domain.ModuleMethod, error)
	UpdateModuleMethod(ctx context.Context, id uint64, input ModuleMethodInput) (domain.ModuleMethod, error)
	DeleteModuleMethod(ctx context.Context, id uint64) error
	ListRolePermissions(ctx context.Context, roleID uint64) ([]domain.RolePermission, error)
	ReplaceRolePermissions(ctx context.Context, roleID uint64, moduleMethodIDs []uint64) error
}

type ModuleMethodInput struct {
	ModuleID    uint64
	Name        string
	Description string
	Method      string
	Path        string
}

type Service struct {
	roleProvider         RoleProvider
	permissionRepository PermissionRepository
	moduleRepository     ModuleRepository
}

func NewService(roleProvider RoleProvider, permissionRepository PermissionRepository, moduleRepository ModuleRepository) *Service {
	return &Service{
		roleProvider:         roleProvider,
		permissionRepository: permissionRepository,
		moduleRepository:     moduleRepository,
	}
}

func (s *Service) SessionFromLoginData(ctx context.Context, data map[string]any) (SessionData, error) {
	user, err := userFromLoginData(data)
	if err != nil {
		return SessionData{}, err
	}

	return s.sessionDataForUser(ctx, user)
}

func (s *Service) SessionForUser(ctx context.Context, user User) (SessionData, error) {
	if user.ID == 0 {
		return SessionData{}, ErrInvalidUser
	}

	return s.sessionDataForUser(ctx, user)
}

func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) {
	if s == nil || s.roleProvider == nil {
		return nil, ErrAuthorizationUnavailable
	}

	roles, err := s.roleProvider.ListRoles(ctx)
	if err != nil {
		return nil, ErrAuthorizationUnavailable
	}

	return roles, nil
}

func (s *Service) RoleHasAccess(ctx context.Context, roleID uint64, method string, path string) (bool, error) {
	if s == nil || s.permissionRepository == nil {
		return false, ErrAuthorizationUnavailable
	}

	return s.permissionRepository.RoleHasAccess(ctx, roleID, method, path)
}

func (s *Service) ListModules(ctx context.Context) ([]domain.Module, error) {
	if s == nil || s.moduleRepository == nil {
		return nil, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.ListModules(ctx)
}

func (s *Service) CreateModule(ctx context.Context, name string) (domain.Module, error) {
	if s == nil || s.moduleRepository == nil {
		return domain.Module{}, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.CreateModule(ctx, strings.TrimSpace(name))
}

func (s *Service) UpdateModule(ctx context.Context, id uint64, name string) (domain.Module, error) {
	if s == nil || s.moduleRepository == nil {
		return domain.Module{}, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.UpdateModule(ctx, id, strings.TrimSpace(name))
}

func (s *Service) DeleteModule(ctx context.Context, id uint64) error {
	if s == nil || s.moduleRepository == nil {
		return ErrAuthorizationUnavailable
	}

	return s.moduleRepository.DeleteModule(ctx, id)
}

func (s *Service) ListModuleMethods(ctx context.Context, moduleID uint64) ([]domain.ModuleMethod, error) {
	if s == nil || s.moduleRepository == nil {
		return nil, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.ListModuleMethods(ctx, moduleID)
}

func (s *Service) CreateModuleMethod(ctx context.Context, input ModuleMethodInput) (domain.ModuleMethod, error) {
	if s == nil || s.moduleRepository == nil {
		return domain.ModuleMethod{}, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.CreateModuleMethod(ctx, normalizeModuleMethodInput(input))
}

func (s *Service) UpdateModuleMethod(ctx context.Context, id uint64, input ModuleMethodInput) (domain.ModuleMethod, error) {
	if s == nil || s.moduleRepository == nil {
		return domain.ModuleMethod{}, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.UpdateModuleMethod(ctx, id, normalizeModuleMethodInput(input))
}

func (s *Service) DeleteModuleMethod(ctx context.Context, id uint64) error {
	if s == nil || s.moduleRepository == nil {
		return ErrAuthorizationUnavailable
	}

	return s.moduleRepository.DeleteModuleMethod(ctx, id)
}

func (s *Service) ListRolePermissions(ctx context.Context, roleID uint64) ([]domain.RolePermission, error) {
	if s == nil || s.moduleRepository == nil {
		return nil, ErrAuthorizationUnavailable
	}

	return s.moduleRepository.ListRolePermissions(ctx, roleID)
}

func (s *Service) ReplaceRolePermissions(ctx context.Context, roleID uint64, moduleMethodIDs []uint64) error {
	if s == nil || s.moduleRepository == nil {
		return ErrAuthorizationUnavailable
	}

	if roleID == 0 {
		return ErrInvalidRole
	}

	return s.moduleRepository.ReplaceRolePermissions(ctx, roleID, moduleMethodIDs)
}

func (s *Service) sessionDataForUser(ctx context.Context, user User) (SessionData, error) {
	permissions := []Permission{}
	if s != nil && s.permissionRepository != nil && user.RoleID > 0 {
		var err error
		permissions, err = s.permissionRepository.ListPermissionsByRoleID(ctx, user.RoleID)
		if err != nil {
			return SessionData{}, ErrAuthorizationUnavailable
		}
	}

	return SessionData{
		UserID:      strconv.FormatUint(user.ID, 10),
		User:        user,
		Permissions: permissions,
	}, nil
}

func userFromLoginData(data map[string]any) (User, error) {
	rawUser, ok := data["user"].(map[string]any)
	if !ok {
		return User{}, ErrInvalidUser
	}

	id, err := uintFromAny(rawUser["id"])
	if err != nil || id == 0 {
		return User{}, ErrInvalidUser
	}

	roleID, _ := uintFromAny(rawUser["role_id"])
	user := User{
		ID:       id,
		FullName: stringFromAny(rawUser["name"]),
		Phone:    stringFromAny(rawUser["phone"]),
		RoleID:   roleID,
		RoleName: stringFromAny(rawUser["role_name"]),
	}

	if sharedauth.IsAdminRole(roleID) {
		return user, nil
	}

	branches, branchIds, err := parseBranches(rawUser["branches"])
	if err != nil {
		return User{}, err
	}

	user.BranchIds = branchIds
	user.Branches = branches

	return user, nil
}

func parseBranches(raw interface{}) ([]Branch, []uint64, error) {
	branchesData, ok := raw.([]interface{})
	if !ok {
		return nil, nil, ErrInvalidUser
	}

	branches := make([]Branch, 0, len(branchesData))
	branchIds := make([]uint64, 0, len(branchesData))
	for _, item := range branchesData {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, nil, ErrInvalidUser
		}

		branches = append(branches, Branch{
			ID:     uint64(m["id"].(float64)),
			KisaAd: m["kisa_ad"].(string),
		})
		branchIds = append(branchIds, uint64(m["id"].(float64)))
	}

	return branches, branchIds, nil
}

func normalizeModuleMethodInput(input ModuleMethodInput) ModuleMethodInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	input.Path = strings.TrimSpace(input.Path)

	return input
}

func uintFromAny(value any) (uint64, error) {
	switch id := value.(type) {
	case float64:
		if id < 0 {
			return 0, ErrInvalidUser
		}

		return uint64(id), nil
	case int:
		if id < 0 {
			return 0, ErrInvalidUser
		}

		return uint64(id), nil
	case uint64:
		return id, nil
	case string:
		return strconv.ParseUint(id, 10, 64)
	default:
		return 0, ErrInvalidUser
	}
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}

	return ""
}
