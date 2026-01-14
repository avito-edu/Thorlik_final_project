package dto

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database,omitempty"`
}
