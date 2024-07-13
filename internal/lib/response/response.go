package response

const (
	StatusOK  = "OK"
	StatusErr = "Error"
)

// Response represents the structure of API responses.
type Response struct {
	Status  string `json:"status"`            // Status of the response (OK or Error)
	Message string `json:"message,omitempty"` // Optional message for successful responses
	Error   string `json:"error,omitempty"`   // Optional error message for error responses
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
