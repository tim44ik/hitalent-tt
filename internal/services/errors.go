package services

const (
	ErrDepartmentNotFound   = "department not found"
	ErrParentNotFound       = "parent department not found"
	ErrEmptyName            = "name cannot be empty"
	ErrNameTooLong          = "name must be at most 200 characters"
	ErrDuplicateName        = "department name already exists under the same parent"
	ErrCyclicMove           = "cannot move department under its own descendant"
	ErrReassignToSelf       = "cannot reassign to the department being deleted"
	ErrReassignToDescendant = "cannot reassign to a descendant department"
)

const (
	ErrEmployeeNotFound  = "employee not found"
	ErrInvalidDepartment = "department does not exist"
	ErrEmptyFullName     = "full name cannot be empty"
	ErrFullNameTooLong   = "full name must be at most 200 characters"
	ErrEmptyPosition     = "position cannot be empty"
	ErrPositionTooLong   = "position must be at most 200 characters"
)
