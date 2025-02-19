package response

import "github.com/gin-gonic/gin"

func makeJsonResponse(
	c *gin.Context,
	status int,
	code string,
	message string,
	payload interface{},
	errors interface{},
	isCached ...bool,
) {
	var res Response
	res.Status = status
	res.Code = code

	// Assign default messages when message is empty
	if message == "" {
		if payload != nil {
			res.Message = "Success"
		} else if errors != nil {
			res.Message = "Failed"
		}
	} else {
		res.Message = message
	}

	// Set data and errors if provided
	if payload != nil {
		res.Data = payload
	}

	if errors != nil {
		res.Errors = errors
	}

	// Handle optional isCached parameter
	if len(isCached) > 0 {
		res.IsCached = isCached[0]
	}

	// Send JSON response
	c.Abort()
	c.JSON(status, res)
}

func Success(c *gin.Context, status int, payload interface{}) {
	makeJsonResponse(c, status, "MSG_SUCCESS", "Success", payload, nil)
}

func Error(c *gin.Context, status int, code, msg string, errors interface{}) {
	if msg == "" {
		makeJsonResponse(c, status, code, "Failed", nil, errors)
	} else {
		makeJsonResponse(c, status, code, msg, nil, errors)
	}
}
