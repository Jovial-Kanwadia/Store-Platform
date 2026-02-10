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

	c.JSON(http.StatusCreated, store)
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

	c.JSON(http.StatusOK, stores)
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

	c.JSON(http.StatusOK, store)
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