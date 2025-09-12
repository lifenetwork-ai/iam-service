package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/infrastructures/rate_limiter/types"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// NormalizeFields normalizes struct fields by setting any field with value "string" to an empty string.
func NormalizeFields(payload interface{}) {
	val := reflect.ValueOf(payload).Elem() // Dereference pointer to get the struct

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Only process string fields
		if field.Kind() == reflect.String {
			// Check if the field is "string" and set it to ""
			if field.String() == "string" {
				field.SetString("")
			}
		}

		// Handle nested structs (if any)
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			NormalizeFields(field.Interface())
		}
	}
}

// ParsePaginationParams parses pagination parameters from the request context.
func ParsePaginationParams(ctx *gin.Context) (int, int, error) {
	page := ctx.DefaultQuery("page", constants.DEFAULT_PAGE)
	size := ctx.DefaultQuery("size", constants.DEFAULT_PAGE_SIZE)

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		return 0, 0, fmt.Errorf("invalid page number")
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		return 0, 0, fmt.Errorf("invalid page size")
	}

	return pageInt, sizeInt, nil
}

// CheckRateLimit checks if the rate limit has been exceeded for a given key.
func CheckRateLimit(
	limiter types.RateLimiter,
	key string,
	maxAttempts int,
	window time.Duration,
) *dto.ErrorDTOResponse {
	limited, err := limiter.IsLimited(key, maxAttempts, window)
	if err != nil {
		logger.GetLogger().Errorf("Rate limiter check failed for key %s: %v", key, err)
		return &dto.ErrorDTOResponse{
			Status:  http.StatusInternalServerError,
			Code:    "MSG_RATE_LIMIT_CHECK_FAILED",
			Message: "Could not check rate limit",
		}
	}
	if limited {
		return &dto.ErrorDTOResponse{
			Status:  http.StatusTooManyRequests,
			Code:    "MSG_RATE_LIMIT_EXCEEDED",
			Message: "Too many attempts, please try again later",
		}
	}
	_ = limiter.RegisterAttempt(key, window)
	return nil
}

// CheckRateLimitDomain checks if the rate limit has been exceeded for a given key.
// Returns a domain error instead of DTO error
func CheckRateLimitDomain(
	limiter types.RateLimiter,
	key string,
	maxAttempts int,
	window time.Duration,
) error {
	limited, err := limiter.IsLimited(key, maxAttempts, window)
	if err != nil {
		logger.GetLogger().Errorf("Rate limiter check failed for key %s: %v", key, err)
		return &domainerrors.DomainError{
			Type:    domainerrors.ErrorTypeInternal,
			Code:    "MSG_RATE_LIMIT_CHECK_FAILED",
			Message: "Could not check rate limit",
			Cause:   err,
		}
	}
	if limited {
		return &domainerrors.DomainError{
			Type:    domainerrors.ErrorTypeRateLimit,
			Code:    "MSG_RATE_LIMIT_EXCEEDED",
			Message: "Too many attempts, please try again later",
		}
	}
	_ = limiter.RegisterAttempt(key, window)
	return nil
}

// ComputeBackoffDuration calculates the backoff duration based on retry count
func ComputeBackoffDuration(retryCount int) time.Duration {
	base := constants.BaseRetryDuration
	maxDelay := constants.DefaultChallengeDuration

	delay := time.Duration(1<<retryCount) * base // 2^retryCount * base
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
