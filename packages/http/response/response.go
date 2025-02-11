package response

import "github.com/gin-gonic/gin"

func makeJsonResponse(
	c *gin.Context,
	status int,
	message string,
	payload interface{},
	errors interface{},
	isCached ...bool,
) {
	var res Response
	res.Status = status
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

func Success(c *gin.Context, status int, msg string, payload interface{}) {
	makeJsonResponse(c, status, msg, payload, nil)
}

func Error(c *gin.Context, status int, msg string, errors interface{}) {
	makeJsonResponse(c, status, msg, nil, errors)
}

// func Errors(c *gin.Context, status int, payload interface{}) {
// 	var res ErrorResponse
// 	res.Status = status
// 	res.Errors = payload
// 	c.Abort()
// 	c.JSON(status, res)
// }

// func NewErrorMap(key string, err error) map[string]interface{} {
// 	res := ErrorMap{
// 		Errors: make(map[string]interface{}),
// 	}
// 	res.Errors[key] = err.Error()
// 	return res.Errors
// }
