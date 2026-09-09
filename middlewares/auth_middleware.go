package middlewares

import (
	"net/http"
	"strings"

	"eventifyApi/config"
	"eventifyApi/utils"

	"github.com/gin-gonic/gin"
)

const (
	RoleAdmin    uint8 = 1
	RolePanitia  uint8 = 2
	RoleCustomer uint8 = 3
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			utils.JSONError(c, http.StatusUnauthorized, "Authorization header is missing", nil)
			c.Abort()
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = strings.TrimSpace(authHeader[7:])
		}

		if tokenStr == "" {
			utils.JSONError(c, http.StatusUnauthorized, "Token is missing in authorization header", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenStr, cfg.JWTSecret)
		if err != nil {
			utils.JSONError(c, http.StatusUnauthorized, "Token is invalid or expired: "+err.Error(), nil)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role_id", claims.RoleID)
		c.Next()
	}
}

func RequireRole(allowedRoles ...uint8) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role_id")
		if !exists {
			utils.JSONError(c, http.StatusUnauthorized, "User context not found", nil)
			c.Abort()
			return
		}

		userRole := roleVal.(uint8)
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.JSONError(c, http.StatusForbidden, "You do not have permission to access this resource", nil)
		c.Abort()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
