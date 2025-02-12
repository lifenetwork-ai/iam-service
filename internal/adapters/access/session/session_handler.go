package access_session

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/dto"
	"github.com/genefriendway/human-network-iam/internal/interfaces"
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
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of sessions"
// @Failure 400 {object} response.GeneralError "Invalid page number or size"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/sessions [get]
func (h *sessionHandler) GetSessions(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")
	keyword := ctx.DefaultQuery("keyword", "")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		logger.GetLogger().Errorf("Invalid page number: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		logger.GetLogger().Errorf("Invalid size: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size"})
		return
	}

	response, errResponse := h.ucase.List(ctx, pageInt, sizeInt, keyword)
	if errResponse != nil {
		logger.GetLogger().Errorf("Failed to get sessions: %v", errResponse)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to get sessions", errResponse)
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
// @Param session_id path string true "session ID"
// @Success 200 {object} dto.sessionDTO "Successful retrieval of session"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "session not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/sessions/{session_id} [get]
func (h *sessionHandler) GetDetail(ctx *gin.Context) {
	// Extract and parse session_id from query string
	sessionId := ctx.Query("session_id")
	if sessionId == "" {
		logger.GetLogger().Error("Invalid session ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	session, err := h.ucase.GetByID(ctx, sessionId)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get session: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, session)
}

// CreateSession creates a new session.
// @Summary Create a new session
// @Description Create a new session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.sessionCreatePayloadDTO true "session payload"
// @Success 201 {object} dto.sessionDTO "Successful creation of session"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/sessions [post]
func (h *sessionHandler) CreateSession(ctx *gin.Context) {
	var reqPayload dto.CreateAccessSessionPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create payment orders, invalid payload", err)
		return
	}

	// Create the session
	response, err := h.ucase.Create(ctx, reqPayload)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create session: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create session", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusCreated, response)
}

// UpdateSession updates an existing session.
// @Summary Update an existing session
// @Description Update an existing session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session_id path string true "session ID"
// @Param session body dto.sessionUpdatePayloadDTO true "session payload"
// @Success 200 {object} dto.sessionDTO "Successful update of session"
// @Failure 400 {object} response.GeneralError "Invalid request payload"
// @Failure 404 {object} response.GeneralError "session not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/sessions/{session_id} [put]
func (h *sessionHandler) UpdateSession(ctx *gin.Context) {

}

// DeleteSession deletes an existing session.
// @Summary Delete an existing session
// @Description Delete an existing session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session_id path string true "session ID"
// @Success 204 "Successful deletion of session"
// @Failure 400 {object} response.GeneralError "Invalid request ID"
// @Failure 404 {object} response.GeneralError "session not found"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/sessions/{session_id} [delete]
func (h *sessionHandler) DeleteSession(ctx *gin.Context) {

}
