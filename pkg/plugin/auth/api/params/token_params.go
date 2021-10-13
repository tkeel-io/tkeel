package params

type TokenCreateReq struct {
	TenantID   string `json:"tenant_id"`
	UserID     string `json:"user_id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}
type TokenCreateResp struct {
	EntityToken string `json:"entity_token"`
}

type TokenParseReq struct {
	EntityToken string `json:"entity_token"`
}
type TokenParseResp struct {
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	TokenID    string `json:"token_id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

type TokenValidReq struct {
	EntityToken string `json:"entity_token"`
}
type TokenValidResp struct {
	IsValid bool `json:"is_valid"`
}
