package auth

const AdminRoleID uint64 = 30

func IsAdminRole(roleID uint64) bool {
	return roleID == AdminRoleID
}
