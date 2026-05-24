package dto

type Response[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

type PageResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

func OK[T any](data T) Response[T] {
	return Response[T]{
		Code:    0,
		Data:    data,
		Message: "ok",
	}
}

func Error(code int, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Data:    nil,
		Message: message,
	}
}
