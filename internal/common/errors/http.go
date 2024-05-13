package errors

type (
	ErrorResponse struct {
		Message string
	}
)

func NewInternalError(err error) *ErrorResponse {
	return &ErrorResponse{Message: err.Error()}
}
