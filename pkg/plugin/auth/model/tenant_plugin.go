package model

// 用户插件
type TenantPlugin struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	PluginID string `json:"plugin_id"`
}
