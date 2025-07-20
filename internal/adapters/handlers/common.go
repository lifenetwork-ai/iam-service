package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	domaintypes "github.com/lifenetwork-ai/iam-service/internal/domain/types"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	httpresponse "github.com/lifenetwork-ai/iam-service/packages/http/response"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

// handleDomainError is a centralized error handler for domain errors
func handleDomainError(ctx *gin.Context, err *domainerrors.DomainError) {
	switch err.Type {
	case domainerrors.ErrorTypeValidation:
		httpresponse.Error(ctx, http.StatusBadRequest, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeNotFound:
		httpresponse.Error(ctx, http.StatusNotFound, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeUnauthorized:
		httpresponse.Error(ctx, http.StatusUnauthorized, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeConflict:
		httpresponse.Error(ctx, http.StatusConflict, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeRateLimit:
		httpresponse.Error(ctx, http.StatusTooManyRequests, err.Code, err.Message, err.Details)
	case domainerrors.ErrorTypeInternal:
		// Log internal errors for debugging
		logger.GetLogger().Errorf("Internal error: %v", err.Error())
		httpresponse.Error(ctx, http.StatusInternalServerError, err.Code, err.Message, err.Details)
	default:
		// Fallback for unknown error types
		logger.GetLogger().Errorf("Unknown error type: %v", err.Error())
		httpresponse.Error(ctx, http.StatusInternalServerError, err.Code, err.Message, err.Details)
	}
}

func ToPaginationDTOResponse[T any](response *domaintypes.PaginatedResponse[T]) *dto.PaginationDTOResponse[T] {
	return &dto.PaginationDTOResponse[T]{
		Items:      response.Items,
		TotalCount: response.TotalCount,
		Page:       response.Page,
		PageSize:   response.PageSize,
		NextPage:   response.NextPage,
	}
}
