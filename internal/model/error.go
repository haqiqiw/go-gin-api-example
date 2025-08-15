package model

const (
	ErrUsernameAlreadyExist = 1000
	ErrUserNotFound         = 1002
	ErrInvalidPassword      = 1003
	ErrInvalidRefreshToken  = 1004
	ErrInvalidLogoutSession = 1005
	ErrInvalidUserID        = 1006
	ErrInvalidOldPassword   = 1007
	ErrInternalServerError  = 1500
)

var errorMessages = map[int]string{
	ErrUsernameAlreadyExist: "Username already exist",
	ErrUserNotFound:         "Username not found",
	ErrInvalidPassword:      "Invalid password",
	ErrInvalidRefreshToken:  "Invalid refresh token",
	ErrInvalidLogoutSession: "Invalid logout session",
	ErrInvalidUserID:        "Invalid user id",
	ErrInvalidOldPassword:   "Invalid old password",
	ErrInternalServerError:  "Internal server error",
}

func MessageFor(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}

type ErrorItem struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CustomError struct {
	HTTPStatus int         `json:"http_status"`
	Errors     []ErrorItem `json:"errors"`
}

func (c *CustomError) Error() string {
	if len(c.Errors) == 0 {
		return "No errors"
	}

	return c.Errors[0].Message
}

func NewCustomError(httpStatus int, codes ...int) *CustomError {
	errors := make([]ErrorItem, len(codes))

	for i, code := range codes {
		errors[i] = ErrorItem{
			Code:    code,
			Message: MessageFor(code),
		}
	}

	return &CustomError{
		HTTPStatus: httpStatus,
		Errors:     errors,
	}
}
