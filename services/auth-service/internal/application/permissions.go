package application

// Permission represents a permission string
type Permission string

// Permission constants
const (
	// User management
	PermUsersCreate   Permission = "users:create"
	PermUsersRead     Permission = "users:read"
	PermUsersUpdate   Permission = "users:update"
	PermUsersDelete   Permission = "users:delete"
	PermUsersAssignRoles Permission = "users:assign_roles"

	// Equipment management
	PermEquipmentCreate  Permission = "equipment:create"
	PermEquipmentRead    Permission = "equipment:read"
	PermEquipmentUpdate  Permission = "equipment:update"
	PermEquipmentDelete  Permission = "equipment:delete"

	// Projects management
	PermProjectsCreate   Permission = "projects:create"
	PermProjectsRead     Permission = "projects:read"
	PermProjectsUpdate   Permission = "projects:update"
	PermProjectsDelete   Permission = "projects:delete"

	// Invoices management
	PermInvoicesCreate   Permission = "invoices:create"
	PermInvoicesRead     Permission = "invoices:read"
	PermInvoicesUpdate   Permission = "invoices:update"
	PermInvoicesDelete   Permission = "invoices:delete"

	// Reports
	PermReportsCreate    Permission = "reports:create"
	PermReportsRead      Permission = "reports:read"
	PermReportsUpdate    Permission = "reports:update"
	PermReportsDelete    Permission = "reports:delete"

	// Settings
	PermSettingsRead     Permission = "settings:read"
	PermSettingsUpdate   Permission = "settings:update"

	// Admin-only
	PermAdminAccess      Permission = "admin:access"
	PermSuperadminAccess Permission = "superadmin:access"
)

// RolePermissions maps roles to their permissions
type RolePermissions map[string][]Permission

// DefaultPermissions defines the default permissions for each role
var DefaultPermissions = RolePermissions{
	"superadmin": {
		// Superadmin has all permissions
		PermSuperadminAccess,
		PermAdminAccess,
		PermUsersCreate,
		PermUsersRead,
		PermUsersUpdate,
		PermUsersDelete,
		PermUsersAssignRoles,
		PermEquipmentCreate,
		PermEquipmentRead,
		PermEquipmentUpdate,
		PermEquipmentDelete,
		PermProjectsCreate,
		PermProjectsRead,
		PermProjectsUpdate,
		PermProjectsDelete,
		PermInvoicesCreate,
		PermInvoicesRead,
		PermInvoicesUpdate,
		PermInvoicesDelete,
		PermReportsCreate,
		PermReportsRead,
		PermReportsUpdate,
		PermReportsDelete,
		PermSettingsRead,
		PermSettingsUpdate,
	},
	"admin": {
		// Admin has all permissions except superadmin-only
		PermAdminAccess,
		PermUsersCreate,
		PermUsersRead,
		PermUsersUpdate,
		PermUsersDelete,
		PermUsersAssignRoles,
		PermEquipmentCreate,
		PermEquipmentRead,
		PermEquipmentUpdate,
		PermEquipmentDelete,
		PermProjectsCreate,
		PermProjectsRead,
		PermProjectsUpdate,
		PermProjectsDelete,
		PermInvoicesCreate,
		PermInvoicesRead,
		PermInvoicesUpdate,
		PermInvoicesDelete,
		PermReportsCreate,
		PermReportsRead,
		PermReportsUpdate,
		PermReportsDelete,
		PermSettingsRead,
		PermSettingsUpdate,
	},
	"manager": {
		// Manager can manage equipment, projects, and view invoices/reports
		PermEquipmentCreate,
		PermEquipmentRead,
		PermEquipmentUpdate,
		PermEquipmentDelete,
		PermProjectsCreate,
		PermProjectsRead,
		PermProjectsUpdate,
		PermProjectsDelete,
		PermInvoicesRead,
		PermReportsRead,
	},
	"warehouse": {
		// Warehouse can read/update equipment and read projects
		PermEquipmentRead,
		PermEquipmentUpdate,
		PermProjectsRead,
	},
	"driver": {
		// Driver can read projects and equipment
		PermProjectsRead,
		PermEquipmentRead,
	},
	"crew": {
		// Crew can read projects and equipment
		PermProjectsRead,
		PermEquipmentRead,
	},
	"freelancer": {
		// Freelancer can read projects and equipment
		PermProjectsRead,
		PermEquipmentRead,
	},
	"readonly": {
		// Read-only access to core resources
		PermEquipmentRead,
		PermProjectsRead,
		PermInvoicesRead,
		PermReportsRead,
	},
}

// HasPermission checks if any of the given roles has the required permission
func HasPermission(roles []string, required Permission) bool {
	for _, role := range roles {
		perms, exists := DefaultPermissions[role]
		if !exists {
			continue
		}
		for _, perm := range perms {
			if perm == required {
				return true
			}
		}
	}
	return false
}

// GetPermissionsForRoles returns all permissions for the given roles (deduplicated)
func GetPermissionsForRoles(roles []string) []Permission {
	permMap := make(map[Permission]bool)
	for _, role := range roles {
		perms, exists := DefaultPermissions[role]
		if !exists {
			continue
		}
		for _, perm := range perms {
			permMap[perm] = true
		}
	}

	permissions := make([]Permission, 0, len(permMap))
	for perm := range permMap {
		permissions = append(permissions, perm)
	}
	return permissions
}
