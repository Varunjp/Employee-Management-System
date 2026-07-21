package response

// ErrorResponse is the JSON body returned for any non-2xx response.
type ErrorResponse struct {
	Message string `json:"message" example:"employee not found"`
} //@name ErrorResponse

func NewError(message string) ErrorResponse {
	return ErrorResponse{Message: message}
}
