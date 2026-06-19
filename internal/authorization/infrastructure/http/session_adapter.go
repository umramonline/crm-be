package http

import (
	"context"

	authhttp "github.com/umran/new.crm/backend/internal/auth/infrastructure/http"
	"github.com/umran/new.crm/backend/internal/authorization/application"
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
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		RoleID:   user.RoleID,
		RoleName: user.RoleName,
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

	return authhttp.SessionData{
		UserID: session.UserID,
		User: authhttp.SessionUser{
			ID:       session.User.ID,
			Name:     session.User.Name,
			Phone:    session.User.Phone,
			RoleID:   session.User.RoleID,
			RoleName: session.User.RoleName,
		},
		Permissions: permissions,
	}
}
