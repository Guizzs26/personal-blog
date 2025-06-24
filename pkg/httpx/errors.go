package httpx

type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeConflict        ErrorCode = "CONFLICT"
	ErrorCodeUnprocessable   ErrorCode = "UNPROCESSABLE_ENTITY"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Server errors (5xx)
	ErrorCodeInternal           ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeBadGateway         ErrorCode = "BAD_GATEWAY"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeGatewayTimeout     ErrorCode = "GATEWAY_TIMEOUT"
)

// APIErrorDetail represents a specific field error (used in validation, for example)
type APIErrorDetail struct {
	Field       string `json:"field,omitempty"`
	Description string `json:"description,omitempty"`
}

// APIError is a structured error response
type APIError struct {
	Error struct {
		Code    ErrorCode        `json:"code"`
		Message string           `json:"message"`
		Status  int              `json:"status"`
		Details []APIErrorDetail `json:"details,omitempty"`
	} `json:"error"`
}
