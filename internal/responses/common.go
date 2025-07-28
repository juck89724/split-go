package responses

// APIResponse 標準 API 回應結構
type APIResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功回應
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Error: false,
		Data:  data,
	}
}

// SuccessWithMessageResponse 帶訊息的成功回應
func SuccessWithMessageResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Error:   false,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 錯誤回應
func ErrorResponse(message string) APIResponse {
	return APIResponse{
		Error:   true,
		Message: message,
	}
}

// PaginationMeta 分頁元資訊
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// PaginatedResponse 分頁回應結構
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// NewPaginatedResponse 創建分頁回應
func NewPaginatedResponse(data interface{}, page, limit int, total int64) PaginatedResponse {
	totalPages := (total + int64(limit) - 1) / int64(limit)

	return PaginatedResponse{
		Data: data,
		Pagination: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
