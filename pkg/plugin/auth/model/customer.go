package model

// 客户.
type Customer struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Address     string `json:"address"`
	CreatedTime int64  `json:"created_time"`
}
