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
	res.Message = message

	if message == "" && payload != nil {
		message = "Success"
	}

	if message == "" && errors != nil {
		message = "Failed"
	}

	if payload != nil {
		res.Data = payload
	}

	if errors != nil {
		res.Errors = errors
	}

	if len(isCached) > 0 {
		res.IsCached = isCached[0]
	}
	c.Abort()
	c.JSON(status, res)
}

func Success(c *gin.Context, status int, payload interface{}) {
	makeJsonResponse(c, status, "MSG_SUCCESS", "Success", payload, nil)
}

func Error(c *gin.Context, status int, code string, msg string, errors interface{}) {
	if msg == "" {
		makeJsonResponse(c, status, code, "Failed", nil, errors)
	} else {
		makeJsonResponse(c, status, code, msg, nil, errors)
	}
}