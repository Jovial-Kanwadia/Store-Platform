package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
)

var (
	dnsNameRegex   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	allowedPlans   = map[string]bool{"small": true, "medium": true}
	allowedEngines = map[string]bool{"woo": true}
)

type StoreService struct {
	repo domain.StoreRepository
}

func NewStoreService(repo domain.StoreRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) CreateStore(ctx context.Context, req domain.CreateStoreRequest) (*domain.Store, error) {
	if err := validateStoreName(req.Name); err != nil {
		return nil, &domain.APIError{Code: 400, Message: err.Error()}
	}

	if !allowedPlans[req.Plan] {
		return nil, &domain.APIError{
			Code:    domain.ErrInvalidPlan.Code,
			Message: fmt.Sprintf("invalid plan %q: allowed values are small, medium", req.Plan),
		}
	}

	if !allowedEngines[req.Engine] {
		return nil, &domain.APIError{
			Code:    domain.ErrInvalidEngine.Code,
			Message: fmt.Sprintf("invalid engine %q: allowed values are woo", req.Engine),
		}
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
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
		Status:    "Pending",
		URL:       fmt.Sprintf("https://%s.stores.example.com", req.Name),
	}

	if err := s.repo.Create(ctx, store); err != nil {
		return nil, &domain.APIError{
			Code:    500,
			Message: "failed to create store",
		}
	}

	return &store, nil
}

func (s *StoreService) ListStores(ctx context.Context, namespace string) ([]domain.Store, error) {
	stores, err := s.repo.List(ctx, namespace)
	if err != nil {
		return nil, &domain.APIError{
			Code:    500,
			Message: "failed to list stores",
		}
	}

	return stores, nil
}

func (s *StoreService) GetStore(ctx context.Context, name, namespace string) (*domain.Store, error) {
	if namespace == "" {
		namespace = "default"
	}

	store, err := s.repo.Get(ctx, name, namespace)
	if err != nil {
		return nil, &domain.APIError{
			Code:    404,
			Message: "store not found",
		}
	}

	return store, nil
}

func (s *StoreService) DeleteStore(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}

	if err := s.repo.Delete(ctx, name, namespace); err != nil {
		return &domain.APIError{
			Code:    500,
			Message: "failed to delete store",
		}
	}

	return nil
}

func validateStoreName(name string) error {
	if name == "" {
		return fmt.Errorf("store name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("store name must be 63 characters or less")
	}

	if !dnsNameRegex.MatchString(strings.ToLower(name)) {
		return fmt.Errorf("store name must be lowercase alphanumeric with hyphens")
	}

	return nil
}