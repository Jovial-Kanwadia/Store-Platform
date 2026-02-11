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