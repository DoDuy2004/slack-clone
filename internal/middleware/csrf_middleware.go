package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware basic CSRF protection for cookie-based auth
// It checks if the Origin or Referer header matches the allowed origins
// and requires a custom header for state-changing requests.
func CSRFMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Skip for safe methods (GET, HEAD, OPTIONS)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 2. Origin/Referer Check
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}

		// If no origin/referer, and it's a state-changing request, it's suspicious
		if origin == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF: Missing Origin or Referer"})
			c.Abort()
			return
		}

		isAllowed := false
		for _, o := range allowedOrigins {
			// Exact match or prefix match for safety
			if origin == o || strings.HasPrefix(origin, o+"/") {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF: Origin not allowed: " + origin})
			c.Abort()
			return
		}

		// 3. Custom Header Check
		// We require "X-CSRF-Token" or "X-Requested-With".
		// Cross-origin requests cannot set these headers without CORS preflight/permission.
		if c.GetHeader("X-Requested-With") == "" && c.GetHeader("X-CSRF-Token") == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF: Missing security headers (X-CSRF-Token or X-Requested-With)"})
			c.Abort()
			return
		}

		c.Next()
	}
}
