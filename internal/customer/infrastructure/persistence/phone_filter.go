package persistence

import (
	"strings"

	"github.com/umran/new.crm/backend/internal/customer/domain"
	"gorm.io/gorm"
)

func applyCustomerPhoneFilter(query *gorm.DB, raw string) *gorm.DB {
	patterns := domain.PhoneLikePatterns(raw)
	if len(patterns) == 0 {
		return query
	}

	clauses := make([]string, 0, len(patterns))
	args := make([]any, 0, len(patterns)*3)

	for _, pattern := range patterns {
		clauses = append(clauses, `(
			cep LIKE ? OR telefon LIKE ? OR EXISTS (
				SELECT 1 FROM customer_telephones ct
				WHERE ct.customer_id = customers.id
					AND ct.deleted_at IS NULL
					AND ct.phone_number LIKE ?
			)
		)`)
		args = append(args, pattern, pattern, pattern)
	}

	return query.Where(strings.Join(clauses, " OR "), args...)
}
