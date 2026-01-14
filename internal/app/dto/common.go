package dto

// ErrorResponse represents error response
// @Description Standard error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message,omitempty" example:"Invalid request body"`
	Code    int    `json:"code,omitempty" example:"400"`
}

// SuccessResponse represents success response
// @Description Standard success response
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationParams represents pagination parameters
// @Description Pagination query parameters
type PaginationParams struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"page_size" example:"20"`
}

// HealthResponse represents health check response
// @Description API health check response
type HealthResponse struct {
	Status   string `json:"status" example:"ok"`
	Service  string `json:"service" example:"marketplace-api"`
	Database string `json:"database,omitempty" example:"connected"`
}
