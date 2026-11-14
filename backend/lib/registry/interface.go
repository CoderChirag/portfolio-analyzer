package registry

import "context"

type RegistryStore[T any] interface {
	GetAll(ctx context.Context) ([]T, error)
}
