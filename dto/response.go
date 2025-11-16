package dto

type ResponseDTO[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func CreateSuccessResponse[T any](message string, data T) ResponseDTO[T] {
	return ResponseDTO[T]{
		Status:  "ok",
		Message: message,
		Data:    data,
	}
}

func CreateErrorResponse(message string) ResponseDTO[any] {
	return ResponseDTO[any]{
		Status:  "error",
		Message: message,
		Data:    nil,
	}
}
