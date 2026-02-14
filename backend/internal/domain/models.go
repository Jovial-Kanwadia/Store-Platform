package domain

import "time"

type Store struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Engine    string    `json:"engine"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateStoreRequest struct {
	Name      string `json:"name" binding:"required"`
	Engine    string `json:"engine" binding:"required"`
	Plan      string `json:"plan" binding:"required"`
	Namespace string `json:"namespace"`
}

type APIError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func (e *APIError) Error() string {
	return e.Message
}

// Sentinel errors for structured HTTP error mapping
var (
	ErrStoreExists   = &APIError{Code: 409, Message: "store already exists"}
	ErrStoreNotFound = &APIError{Code: 404, Message: "store not found"}
	ErrInvalidName   = &APIError{Code: 400, Message: "invalid store name"}
	ErrInvalidPlan   = &APIError{Code: 400, Message: "invalid plan"}
	ErrInvalidEngine = &APIError{Code: 400, Message: "invalid engine"}
	ErrInternal      = &APIError{Code: 500, Message: "internal server error"}
)
