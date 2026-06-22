package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/umran/new.crm/backend/internal/authorization/application"
	"github.com/umran/new.crm/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

type moduleRequest struct {
	Name string `json:"name"`
}

type moduleMethodRequest struct {
	ModuleID    uint64 `json:"module_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Path        string `json:"path"`
}

type replaceRolePermissionsRequest struct {
	ModuleMethodIDs []uint64 `json:"module_method_ids"`
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authRequired fiber.Handler) {
	router.Get("/authorization/roles", authRequired, h.ListRoles)
	router.Get("/authorization/modules", authRequired, h.ListModules)
	router.Post("/authorization/modules", authRequired, h.CreateModule)
	router.Put("/authorization/modules/:id", authRequired, h.UpdateModule)
	router.Delete("/authorization/modules/:id", authRequired, h.DeleteModule)
	router.Get("/authorization/module-methods", authRequired, h.ListModuleMethods)
	router.Post("/authorization/module-methods", authRequired, h.CreateModuleMethod)
	router.Put("/authorization/module-methods/:id", authRequired, h.UpdateModuleMethod)
	router.Delete("/authorization/module-methods/:id", authRequired, h.DeleteModuleMethod)
	router.Get("/authorization/role-permissions", authRequired, h.ListRolePermissions)
	router.Put("/authorization/role-permissions/:role_id", authRequired, h.ReplaceRolePermissions)
}

func (h *Handler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.service.ListRoles(c.UserContext())
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Roller getirildi.", fiber.Map{"items": roles})
}

func (h *Handler) ListModules(c *fiber.Ctx) error {
	modules, err := h.service.ListModules(c.UserContext())
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modüller getirildi.", fiber.Map{"items": modules})
}

func (h *Handler) CreateModule(c *fiber.Ctx) error {
	var request moduleRequest
	if err := c.BodyParser(&request); err != nil {
		return invalidBody(c)
	}

	module, err := h.service.CreateModule(c.UserContext(), request.Name)
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusCreated, "Modül oluşturuldu.", module)
}

func (h *Handler) UpdateModule(c *fiber.Ctx) error {
	id, err := paramUint(c, "id")
	if err != nil {
		return invalidID(c)
	}

	var request moduleRequest
	if err := c.BodyParser(&request); err != nil {
		return invalidBody(c)
	}

	module, err := h.service.UpdateModule(c.UserContext(), id, request.Name)
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modül güncellendi.", module)
}

func (h *Handler) DeleteModule(c *fiber.Ctx) error {
	id, err := paramUint(c, "id")
	if err != nil {
		return invalidID(c)
	}

	if err := h.service.DeleteModule(c.UserContext(), id); err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modül silindi.", fiber.Map{})
}

func (h *Handler) ListModuleMethods(c *fiber.Ctx) error {
	moduleID, err := queryUint(c, "module_id")
	if err != nil {
		return invalidID(c)
	}

	methods, err := h.service.ListModuleMethods(c.UserContext(), moduleID)
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modül methodları getirildi.", fiber.Map{"items": methods})
}

func (h *Handler) CreateModuleMethod(c *fiber.Ctx) error {
	var request moduleMethodRequest
	if err := c.BodyParser(&request); err != nil {
		return invalidBody(c)
	}

	method, err := h.service.CreateModuleMethod(c.UserContext(), moduleMethodInput(request))
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusCreated, "Modül methodu oluşturuldu.", method)
}

func (h *Handler) UpdateModuleMethod(c *fiber.Ctx) error {
	id, err := paramUint(c, "id")
	if err != nil {
		return invalidID(c)
	}

	var request moduleMethodRequest
	if err := c.BodyParser(&request); err != nil {
		return invalidBody(c)
	}

	method, err := h.service.UpdateModuleMethod(c.UserContext(), id, moduleMethodInput(request))
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modül methodu güncellendi.", method)
}

func (h *Handler) DeleteModuleMethod(c *fiber.Ctx) error {
	id, err := paramUint(c, "id")
	if err != nil {
		return invalidID(c)
	}

	if err := h.service.DeleteModuleMethod(c.UserContext(), id); err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Modül methodu silindi.", fiber.Map{})
}

func (h *Handler) ListRolePermissions(c *fiber.Ctx) error {
	roleID, err := queryUint(c, "role_id")
	if err != nil || roleID == 0 {
		return invalidID(c)
	}

	permissions, err := h.service.ListRolePermissions(c.UserContext(), roleID)
	if err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Rol izinleri getirildi.", fiber.Map{"items": permissions})
}

func (h *Handler) ReplaceRolePermissions(c *fiber.Ctx) error {
	roleID, err := paramUint(c, "role_id")
	if err != nil || roleID == 0 {
		return invalidID(c)
	}

	var request replaceRolePermissionsRequest
	if err := c.BodyParser(&request); err != nil {
		return invalidBody(c)
	}

	if err := h.service.ReplaceRolePermissions(c.UserContext(), roleID, request.ModuleMethodIDs); err != nil {
		return h.serviceUnavailable(c)
	}

	return response.Success(c, fiber.StatusOK, "Rol izinleri güncellendi.", fiber.Map{})
}

func (h *Handler) serviceUnavailable(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusServiceUnavailable, "Yetkilendirme servisi şu anda kullanılamıyor.", nil)
}

func invalidBody(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz istek gövdesi.", map[string]string{
		"body": "JSON formatı geçersiz.",
	})
}

func invalidID(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusUnprocessableEntity, "Geçersiz kayıt bilgisi.", nil)
}

func paramUint(c *fiber.Ctx, name string) (uint64, error) {
	return strconv.ParseUint(c.Params(name), 10, 64)
}

func queryUint(c *fiber.Ctx, name string) (uint64, error) {
	value := c.Query(name)
	if value == "" {
		return 0, nil
	}

	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func moduleMethodInput(request moduleMethodRequest) application.ModuleMethodInput {
	return application.ModuleMethodInput{
		ModuleID:    request.ModuleID,
		Name:        request.Name,
		Description: request.Description,
		Method:      request.Method,
		Path:        request.Path,
	}
}
