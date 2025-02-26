package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	interfaces "github.com/genefriendway/human-network-iam/internal/domain/ucases/types"
	httpresponse "github.com/genefriendway/human-network-iam/packages/http/response"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

type sessionHandler struct {
	ucase interfaces.AccessSessionUseCase
}

func NewAccessSessionHandler(ucase interfaces.AccessSessionUseCase) *sessionHandler {
	return &sessionHandler{
		ucase: ucase,
	}
}

// GetSessions retrieves a list of sessions.
// @Summary Retrieve sessions
// @Description Get sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of sessions"
// @Failure 400 {object} response.ErrorResponse "Invalid page number or size"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/sessions [get]
func (h *sessionHandler) GetSessions(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")
	keyword := ctx.DefaultQuery("keyword", "")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.GetLogger().Errorf("Invalid page number: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_PAGE_NUMBER",
			"Invalid page number",
			err,
		)
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		logger.GetLogger().Errorf("Invalid size: %v", err)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_INVALID_SIZE",
			"Invalid size",
			err,
		)
		return
	}

	response, errResponse := h.ucase.List(ctx, pageInt, sizeInt, keyword)
	if errResponse != nil {
		logger.GetLogger().Errorf("Failed to get sessions: %v", errResponse)
		httpresponse.Error(
			ctx,
			http.StatusBadRequest,
			"MSG_FAILED_GET_SESSIONS",
			"Failed to get sessions",
			errResponse,
		)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}

// GetDetail retrieves a session by it's ID.
// @Summary Retrieve session by ID
// @Description Get session by ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param session_id path string true "session ID"
// @Success 200 {object} dto.AccessSessionDTO "Successful retrieval of session"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/sessions/{session_id} [get]
func (h *sessionHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse session_id from query string
	sessionId := ctx.Query("session_id")
	if sessionId == "" {
		logger.GetLogger().Error("Invalid session ID")
		httpresponse.Error(ctx, http.StatusBadRequest, "MSG_INVALID_SESSION_ID", "Invalid session ID", nil)
		return
	}

	session, err := h.ucase.GetByID(ctx, sessionId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		return
	}

	// Return the response as a JSON response
	httpresponse.Success(ctx, http.StatusOK, session)
}

// DeleteSession deletes an existing session.
// @Summary Delete an existing session
// @Description Delete an existing session
// @Tags sessions
// @Accept json
// @Produce json
// @Param X-Organization-Id header string true "Organization ID"
// @Param Authorization header string true "Bearer Token"
// @Param session_id path string true "session ID"
// @Success 204 "Successful deletion of session"
// @Failure 400 {object} response.ErrorResponse "Invalid request ID"
// @Failure 404 {object} response.ErrorResponse "session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/sessions/{session_id} [delete]
func (h *sessionHandler) DeleteSession(ctx *gin.Context) {
	httpresponse.Error(
		ctx,
		http.StatusNotImplemented,
		"MSG_NOT_IMPLEMENTED",
		"Not implemented",
		nil,
	)
}
