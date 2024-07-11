package response

const (
	StatusOK  = "OK"
	StatusErr = "Error"
)

// Response - общий ответ API
type Response struct {
	Status  string `json:"status"`            // Статус ответа
	Message string `json:"message,omitempty"` // Сообщение, если есть
	Error   string `json:"error,omitempty"`   // Ошибка, если есть
}

// Ok - функция для создания успешного ответа
func Ok(msg string) Response {
	return Response{
		Status:  StatusOK,
		Message: msg,
	}
}

// Err - функция для создания ответа с ошибкой
func Err(errMsg string) Response {
	return Response{
		Status: StatusErr,
		Error:  errMsg,
	}
}
