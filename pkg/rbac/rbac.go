package rbac

import "context"

type Operation uint

const (
	READ Operation = iota
	UPDATE
	CREATE
	DELETE
)

type Guarder[R, O any] func(context.Context, Operation, R, O) (bool, error)

type rbac[R, O any] struct {
	guarder Guarder[R, O]
}

func NewRBAC[R, O any](guarder Guarder[R, O]) *rbac[R, O] {
	return &rbac[R, O]{
		guarder: guarder,
	}
}

func (r *rbac[R, O]) IsGranted(ctx context.Context, op Operation, resource R, owner O) (bool, error) {
	return r.guarder(ctx, op, resource, owner)
}

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
