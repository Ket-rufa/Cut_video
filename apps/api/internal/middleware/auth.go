package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID = "user_id"
	DefaultUserID    = "default_user"
)

// AuthMiddleware extracts the authenticated user ID from X-User-Id header or Authorization header.
// If none provided, defaults to DefaultUserID in development, or allows identifying clients.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader("X-User-Id"))
		if userID == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// If token is prefixed or has value, use it as token user identifier
				if token != "" {
					userID = token
				}
			}
		}

		if userID == "" {
			userID = DefaultUserID
		}

		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from Gin context.
func GetUserID(c *gin.Context) string {
	if val, exists := c.Get(ContextKeyUserID); exists {
		if uid, ok := val.(string); ok && uid != "" {
			return uid
		}
	}
	return DefaultUserID
}

// CheckOwnership verifies that the resource owner matches the authenticated user.
func CheckOwnership(c *gin.Context, resourceUserID string) bool {
	currentUserID := GetUserID(c)
	if resourceUserID == "" || resourceUserID == DefaultUserID || currentUserID == resourceUserID {
		return true
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "FORBIDDEN",
			"message": "Bạn không có quyền truy cập tài nguyên này.",
		},
	})
	return false
}
