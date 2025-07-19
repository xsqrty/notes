package rbac

import "context"

// Operation represents an action or permission type, typically used within RBAC (Role-Based Access Control) systems.
type Operation uint

const (
	// READ represents a read operation in the Operation enum.
	READ Operation = iota
	// UPDATE represents an update operation in the Operation enum.
	UPDATE
	// CREATE represents a create operation in the Operation enum.
	CREATE
	// DELETE represents a delete operation in the Operation enum.
	DELETE
)

// Guarder represents a function that evaluates permissions based on context, operation, resource, and options.
// Returns a boolean indicating access permission and an error if evaluation fails.
type Guarder[R, O any] func(context.Context, Operation, R, O) (bool, error)

// rbac is a type used to implement role-based access control logic with customizable guard functions.
type rbac[R, O any] struct {
	guarder Guarder[R, O]
}

// NewRBAC initializes a new RBAC instance with the provided Guarder function and returns it.
func NewRBAC[R, O any](guarder Guarder[R, O]) *rbac[R, O] {
	return &rbac[R, O]{
		guarder: guarder,
	}
}

// IsGranted evaluates if the given operation on a resource is permitted for the specified owner, using the defined guarder.
func (r *rbac[R, O]) IsGranted(ctx context.Context, op Operation, resource R, owner O) (bool, error) {
	return r.guarder(ctx, op, resource, owner)
}

// String returns the string representation of the Operation type.
func (o Operation) String() string {
	switch o {
	case READ:
		return "read"
	case UPDATE:
		return "update"
	case CREATE:
		return "create"
	case DELETE:
		return "delete"
	}

	return "unknown"
}
