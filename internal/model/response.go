package model

type Meta struct {
	HTTPStatus int `json:"http_status"`
}

type MetaWithPage struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	Total      int `json:"total"`
	HTTPStatus int `json:"http_status"`
}

type SuccessResponse[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

type SuccessListResponse[T any] struct {
	Data         []T          `json:"data"`
	MetaWithPage MetaWithPage `json:"meta"`
}

type SuccessMessageResponse struct {
	Message string `json:"message"`
	Meta    Meta   `json:"meta"`
}

type ErrorResponse struct {
	Errors []ErrorItem `json:"errors"`
	Meta   Meta        `json:"meta"`
}

func NewSuccessResponse[T any](data T, httpStatus int) SuccessResponse[T] {
	return SuccessResponse[T]{
		Data: data,
		Meta: Meta{
			HTTPStatus: httpStatus,
		},
	}
}

func NewSuccessListResponse[T any](data []T, meta MetaWithPage) SuccessListResponse[T] {
	return SuccessListResponse[T]{
		Data:         data,
		MetaWithPage: meta,
	}
}

func NewSuccessMessageResponse(msg string, httpStatus int) SuccessMessageResponse {
	return SuccessMessageResponse{
		Message: msg,
		Meta: Meta{
			HTTPStatus: httpStatus,
		},
	}
}
