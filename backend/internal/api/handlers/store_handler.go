package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/service"
)

type StoreHandler struct {
	svc *service.StoreService
}

func NewStoreHandler(svc *service.StoreService) *StoreHandler {
	return &StoreHandler{svc: svc}
}

type storeResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Engine    string `json:"engine"`
	Plan      string `json:"plan"`
	Status    string `json:"status"`
	URL       string `json:"url,omitempty"`
	CreatedAt string `json:"createdAt"`
}

func toStoreResponse(s domain.Store) storeResponse {
	return storeResponse{
		Name:      s.Name,
		Namespace: s.Namespace,
		Engine:    s.Engine,
		Plan:      s.Plan,
		Status:    s.Status,
		URL:       s.URL,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *StoreHandler) Create(c *gin.Context) {
	var req domain.CreateStoreRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	store, err := h.svc.CreateStore(c.Request.Context(), req)
	if err != nil {
		if apiErr, ok := err.(*domain.APIError); ok {
			c.JSON(apiErr.Code, gin.H{
				"error": apiErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, toStoreResponse(*store))
}

func (h *StoreHandler) List(c *gin.Context) {
	namespace := c.Query("namespace")

	stores, err := h.svc.ListStores(c.Request.Context(), namespace)
	if err != nil {
		if apiErr, ok := err.(*domain.APIError); ok {
			c.JSON(apiErr.Code, gin.H{
				"error": apiErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	resp := make([]storeResponse, 0, len(stores))
	for _, s := range stores {
		resp = append(resp, toStoreResponse(s))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *StoreHandler) Get(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Query("namespace")

	store, err := h.svc.GetStore(c.Request.Context(), name, namespace)
	if err != nil {
		if apiErr, ok := err.(*domain.APIError); ok {
			c.JSON(apiErr.Code, gin.H{
				"error": apiErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, toStoreResponse(*store))
}

func (h *StoreHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Query("namespace")

	err := h.svc.DeleteStore(c.Request.Context(), name, namespace)
	if err != nil {
		if apiErr, ok := err.(*domain.APIError); ok {
			c.JSON(apiErr.Code, gin.H{
				"error": apiErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}