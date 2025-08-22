package model

import "net/http"

var (
	ErrInternalServerError        = NewCustomError(http.StatusInternalServerError, 100, "internal server error")
	ErrUnauthorized               = NewCustomError(http.StatusUnauthorized, 101, "unauthorized")
	ErrBadRequest                 = NewCustomError(http.StatusBadRequest, 102, "bad request")
	ErrForbidden                  = NewCustomError(http.StatusForbidden, 103, "forbidden")
	ErrMissingOrInvalidAuthHeader = NewCustomError(http.StatusUnauthorized, 104, "missing or invalid auth header")
	ErrInvalidAuthToken           = NewCustomError(http.StatusUnauthorized, 105, "invalid auth token")
	ErrTokenRevoked               = NewCustomError(http.StatusUnauthorized, 106, "token revoked")

	ErrUsernameAlreadyExist = NewCustomError(http.StatusBadRequest, 1000, "username already exist")
	ErrUserNotFound         = NewCustomError(http.StatusNotFound, 1002, "username not found")
	ErrInvalidPassword      = NewCustomError(http.StatusUnauthorized, 1003, "invalid password")
	ErrInvalidRefreshToken  = NewCustomError(http.StatusUnauthorized, 1004, "invalid refresh token")
	ErrInvalidLogoutSession = NewCustomError(http.StatusUnauthorized, 1005, "invalid logout session")
	ErrInvalidUserID        = NewCustomError(http.StatusUnprocessableEntity, 1006, "invalid user id")
	ErrInvalidOldPassword   = NewCustomError(http.StatusBadRequest, 1007, "invalid old password")
)

type ErrorItem struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CustomError struct {
	HTTPStatus int         `json:"http_status"`
	Errors     []ErrorItem `json:"errors"`
}

func (c *CustomError) Append(ei ErrorItem) {
	c.Errors = append(c.Errors, ei)
}

func (c *CustomError) Error() string {
	if len(c.Errors) == 0 {
		return "empty error"
	}

	return c.Errors[0].Message
}

func NewCustomError(httpStatus int, code int, msg string) *CustomError {
	return &CustomError{
		HTTPStatus: httpStatus,
		Errors: []ErrorItem{
			{
				Code:    code,
				Message: msg,
			},
		},
	}
}
