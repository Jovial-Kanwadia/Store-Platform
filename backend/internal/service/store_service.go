package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/config"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
)

var dnsNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type StoreService struct {
	repo domain.StoreRepository
	cfg  *config.Config
}

func NewStoreService(repo domain.StoreRepository, cfg *config.Config) *StoreService {
	return &StoreService{repo: repo, cfg: cfg}
}

func (s *StoreService) CreateStore(ctx context.Context, req domain.CreateStoreRequest) (*domain.Store, error) {
	if err := validateStoreName(req.Name); err != nil {
		return nil, &domain.APIError{Code: domain.ErrInvalidName.Code, Message: err.Error()}
	}

	if !domain.AllowedPlans[req.Plan] {
		return nil, &domain.APIError{
			Code:    domain.ErrInvalidPlan.Code,
			Message: fmt.Sprintf("invalid plan %q: allowed values are small, medium, large", req.Plan),
		}
	}

	if !domain.AllowedEngines[req.Engine] {
		return nil, &domain.APIError{
			Code:    domain.ErrInvalidEngine.Code,
			Message: fmt.Sprintf("invalid engine %q: allowed values are woo", req.Engine),
		}
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = domain.DefaultNamespace
	}

	// Duplicate check
	existing, _ := s.repo.Get(ctx, strings.ToLower(req.Name), namespace)
	if existing != nil {
		return nil, &domain.APIError{
			Code:    domain.ErrStoreExists.Code,
			Message: fmt.Sprintf("store %q already exists", req.Name),
		}
	}

	store := domain.Store{
		Name:      strings.ToLower(req.Name),
		Namespace: namespace,
		Engine:    req.Engine,
		Plan:      req.Plan,
		Status:    domain.StatusPending,
		URL:       fmt.Sprintf("https://%s.%s", req.Name, s.cfg.BaseDomain),
	}

	if err := s.repo.Create(ctx, store); err != nil {
		return nil, &domain.APIError{
			Code:    domain.ErrInternal.Code,
			Message: "failed to create store",
		}
	}

	return &store, nil
}

func (s *StoreService) ListStores(ctx context.Context, namespace string) ([]domain.Store, error) {
	stores, err := s.repo.List(ctx, namespace)
	if err != nil {
		return nil, &domain.APIError{
			Code:    domain.ErrInternal.Code,
			Message: "failed to list stores",
		}
	}

	return stores, nil
}

func (s *StoreService) GetStore(ctx context.Context, name, namespace string) (*domain.Store, error) {
	if namespace == "" {
		namespace = domain.DefaultNamespace
	}

	store, err := s.repo.Get(ctx, name, namespace)
	if err != nil {
		return nil, &domain.APIError{
			Code:    domain.ErrStoreNotFound.Code,
			Message: "store not found",
		}
	}

	return store, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = domain.DefaultNamespace
	}

	if err := s.repo.Delete(ctx, name, namespace); err != nil {
		return &domain.APIError{
			Code:    domain.ErrInternal.Code,
			Message: "failed to delete store",
		}
	}

	return nil
}

func validateStoreName(name string) error {
	if name == "" {
		return fmt.Errorf("store name cannot be empty")
	}

	if len(name) > domain.MaxStoreNameLength {
		return fmt.Errorf("store name must be %d characters or less", domain.MaxStoreNameLength)
	}

	if !dnsNameRegex.MatchString(strings.ToLower(name)) {
		return fmt.Errorf("store name must be lowercase alphanumeric with hyphens")
	}

	return nil
}
