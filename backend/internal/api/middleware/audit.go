package middleware

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditLogger logs mutating API actions (POST, DELETE) with structured fields.
func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "DELETE" {
			c.Next()
			return
		}

		// Determine action and store name from path
		path := c.Request.URL.Path
		action := "unknown"
		storeName := ""

		switch {
		case method == "POST" && strings.HasSuffix(path, "/stores"):
			action = "create_store"
		case method == "DELETE" && strings.Contains(path, "/stores/"):
			action = "delete_store"
			storeName = c.Param("name")
		}

		c.Next()

		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}

		slog.Info("audit",
			"event", "audit",
			"action", action,
			"store", storeName,
			"ip", c.ClientIP(),
			"status", status,
			"http_status", c.Writer.Status(),
		)
	}
}
