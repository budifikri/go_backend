package response

// ApiResponse standard API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse for list endpoints
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}

// Pagination metadata
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// ErrorResponse for error responses
type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationError for validation failures
type ValidationError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Field   string `json:"field,omitempty"`
	Value   any    `json:"value,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}, message string) ApiResponse {
	return ApiResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(errorMsg string) ApiResponse {
	return ApiResponse{
		Success: false,
		Error:   errorMsg,
	}
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, total int64, limit, offset int) PaginatedResponse {
	hasMore := int64(offset+limit) < total
	return PaginatedResponse{
		Success: true,
		Data:    data,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}
