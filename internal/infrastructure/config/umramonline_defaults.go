package config

import "strings"

// Defaults match umramonline routes/api.php (Admin + CRM data prefix).
const (
	defaultUmramonlineOTPRequestPath          = "/api/v1/admin/login"
	defaultUmramonlineOTPVerifyPath           = "/api/v1/admin/login/verify"
	defaultUmramonlinePasswordLoginPath         = "/api/v1/admin/login"
	defaultUmramonlineUserRolesPath           = "/api/v1/admin/crm/user-roles"
	defaultUmramonlineCustomersPath           = "/api/v1/admin/crm/customer-list"
	defaultUmramonlineCustomerSearchPath      = "/api/v1/admin/customer/search"
	defaultUmramonlineCustomerPhoneExistsPath = ""
	defaultUmramonlineZonesPath               = "/api/v1/admin/crm/zones"
	defaultUmramonlineCitiesPath              = "/api/v1/admin/cities"
	defaultUmramonlineTownsPath               = "/api/v1/admin/towns"
	defaultUmramonlineBranchesPath            = "/api/v1/admin/crm/branches"
	defaultUmramonlineTaskSMSPath             = ""
	defaultUmramonlineDashboardPath           = "/api/v1/admin/dashboard"
)

func (c *Config) applyUmramonlineDefaults() {
	c.UmramonlineOTPRequestPath = defaultPath(c.UmramonlineOTPRequestPath, defaultUmramonlineOTPRequestPath)
	c.UmramonlineOTPVerifyPath = defaultPath(c.UmramonlineOTPVerifyPath, defaultUmramonlineOTPVerifyPath)
	c.UmramonlinePasswordPath = defaultPath(c.UmramonlinePasswordPath, defaultUmramonlinePasswordLoginPath)
	c.UmramonlineUserRolesPath = defaultPath(c.UmramonlineUserRolesPath, defaultUmramonlineUserRolesPath)
	c.UmramonlineCustomersPath = defaultPath(c.UmramonlineCustomersPath, defaultUmramonlineCustomersPath)
	c.UmramonlineCustomerSearchPath = defaultPath(c.UmramonlineCustomerSearchPath, defaultUmramonlineCustomerSearchPath)
	c.UmramonlineCustomerPhoneExistsPath = defaultOptionalPath(c.UmramonlineCustomerPhoneExistsPath, defaultUmramonlineCustomerPhoneExistsPath)
	c.UmramonlineZonesPath = defaultPath(c.UmramonlineZonesPath, defaultUmramonlineZonesPath)
	c.UmramonlineCitiesPath = defaultPath(c.UmramonlineCitiesPath, defaultUmramonlineCitiesPath)
	c.UmramonlineTownsPath = defaultPath(c.UmramonlineTownsPath, defaultUmramonlineTownsPath)
	c.UmramonlineBranchesPath = defaultPath(c.UmramonlineBranchesPath, defaultUmramonlineBranchesPath)
	c.UmramonlineTaskSMSPath = defaultOptionalPath(c.UmramonlineTaskSMSPath, defaultUmramonlineTaskSMSPath)
	c.UmramonlineDashboardPath = defaultPath(c.UmramonlineDashboardPath, defaultUmramonlineDashboardPath)
}

func defaultPath(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}

	return fallback
}

// defaultOptionalPath keeps env empty when no fallback is configured (feature disabled in client).
func defaultOptionalPath(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}

	return fallback
}
