package domain

import "context"

type StoreRepository interface {
	Create(ctx context.Context, s Store) error
	List(ctx context.Context, namespace string) ([]Store, error)
	Get(ctx context.Context, name, namespace string) (*Store, error)
	Delete(ctx context.Context, name, namespace string) error
}

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}