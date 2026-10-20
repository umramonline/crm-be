package config

import "testing"

func TestApplyUmramonlineDefaultsUsesAdminPaths(t *testing.T) {
	cfg := &Config{}
	cfg.applyUmramonlineDefaults()

	if cfg.UmramonlineDashboardPath != "/api/v1/admin/dashboard" {
		t.Fatalf("unexpected dashboard path: %q", cfg.UmramonlineDashboardPath)
	}

	if cfg.UmramonlineCustomerSearchPath != "/api/v1/admin/customer/search" {
		t.Fatalf("unexpected customer search path: %q", cfg.UmramonlineCustomerSearchPath)
	}

	if cfg.UmramonlineCustomerPhoneExistsPath != "" {
		t.Fatalf("expected empty phone-exists path, got %q", cfg.UmramonlineCustomerPhoneExistsPath)
	}
}

func TestApplyUmramonlineDefaultsPreservesExplicitPaths(t *testing.T) {
	cfg := &Config{
		UmramonlineDashboardPath: " /api/v1/custom/dashboard ",
	}
	cfg.applyUmramonlineDefaults()

	if cfg.UmramonlineDashboardPath != "/api/v1/custom/dashboard" {
		t.Fatalf("expected trimmed explicit path, got %q", cfg.UmramonlineDashboardPath)
	}
}

func TestApplyUmramonlineDefaultsFillsBlankEnvValues(t *testing.T) {
	cfg := &Config{
		UmramonlineDashboardPath: "   ",
		UmramonlineTaskSMSPath:   " ",
	}
	cfg.applyUmramonlineDefaults()

	if cfg.UmramonlineDashboardPath != defaultUmramonlineDashboardPath {
		t.Fatalf("expected dashboard fallback, got %q", cfg.UmramonlineDashboardPath)
	}

	if cfg.UmramonlineTaskSMSPath != "" {
		t.Fatalf("expected empty task sms path, got %q", cfg.UmramonlineTaskSMSPath)
	}
}
