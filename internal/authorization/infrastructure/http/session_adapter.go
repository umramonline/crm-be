package http

import (
	"context"

	authhttp "github.com/umran/new.crm/backend/internal/auth/infrastructure/http"
	"github.com/umran/new.crm/backend/internal/authorization/application"
	sharedauth "github.com/umran/new.crm/backend/internal/shared/auth"
)

type SessionAdapter struct {
	service *application.Service
}

func NewSessionAdapter(service *application.Service) *SessionAdapter {
	return &SessionAdapter{service: service}
}

func (a *SessionAdapter) SessionFromLoginData(ctx context.Context, data map[string]any) (authhttp.SessionData, error) {
	session, err := a.service.SessionFromLoginData(ctx, data)
	if err != nil {
		return authhttp.SessionData{}, err
	}

	return toAuthSessionData(session), nil
}

func (a *SessionAdapter) SessionForUser(ctx context.Context, user authhttp.SessionUser) (authhttp.SessionData, error) {
	session, err := a.service.SessionForUser(ctx, application.User{
		ID:        user.ID,
		FullName:  user.FullName,
		Phone:     user.Phone,
		BranchIds: user.BranchIds,
		RoleID:    user.RoleID,
		RoleName:  user.RoleName,
	})
	if err != nil {
		return authhttp.SessionData{}, err
	}

	return toAuthSessionData(session), nil
}

func toAuthSessionData(session application.SessionData) authhttp.SessionData {
	permissions := make([]authhttp.Permission, 0, len(session.Permissions))
	for _, permission := range session.Permissions {
		permissions = append(permissions, authhttp.Permission{
			ModuleID:       permission.ModuleID,
			ModuleName:     permission.ModuleName,
			ModuleMethodID: permission.ModuleMethodID,
			Name:           permission.Name,
			Description:    permission.Description,
			Method:         permission.Method,
			Path:           permission.Path,
		})
	}

	user := authhttp.SessionUser{
		ID:       session.User.ID,
		FullName: session.User.FullName,
		Phone:    session.User.Phone,
		RoleID:   session.User.RoleID,
		RoleName: session.User.RoleName,
	}

	if !sharedauth.IsAdminRole(session.User.RoleID) {
		user.BranchIds = session.User.BranchIds
		user.Branches = session.User.Branches
	}

	return authhttp.SessionData{
		UserID:      session.User.ID,
		User:        user,
		Permissions: permissions,
	}
}
