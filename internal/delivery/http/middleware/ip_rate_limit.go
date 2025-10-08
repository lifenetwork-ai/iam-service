package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ratelimiters "github.com/lifenetwork-ai/iam-service/infrastructures/rate_limiter/types"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// RateLimitConfig defines middleware configuration.
type RateLimitConfig struct {
	RateLimiter ratelimiters.RateLimiter
	Action      string        // e.g. "login_phone", "login_email"
	Limit       int           // e.g. 5
	Window      time.Duration // e.g. 5 * time.Minute
}

// IPRateLimitMiddleware returns a Gin middleware to enforce rate-limiting per IP and tenant.
func IPRateLimitMiddleware(cfg RateLimitConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cfg.RateLimiter == nil {
			ctx.Next()
			return
		}

		clientIP := ctx.ClientIP()
		if clientIP == "" {
			clientIP = "unknown"
		}

		logger.GetLogger().Infof("IP Rate Limiting: action=%s, limit=%d, window=%s, client_ip=%s",
			cfg.Action, cfg.Limit, cfg.Window.String(), clientIP)

		tenant, _ := GetTenantFromContext(ctx)
		tenantID := "unknown"
		if tenant != nil {
			tenantID = tenant.ID.String()
		}

		key := fmt.Sprintf("rl:%s:%s:%s", tenantID, cfg.Action, clientIP)

		limited, err := cfg.RateLimiter.IsLimited(key, cfg.Limit, cfg.Window)
		if err != nil {
			ctx.Next() // fail-open if cache error
			return
		}

		if limited {
			httpresponse.Error(
				ctx,
				http.StatusTooManyRequests,
				"MSG_RATE_LIMIT",
				fmt.Sprintf("Too many OTP requests from IP %s. Please try again later.", clientIP),
				nil,
			)
			ctx.Abort()
			return
		}

		// Record new attempt
		_ = cfg.RateLimiter.RegisterAttempt(key, cfg.Window)

		ctx.Next()
	}
}
