// internal/tenant/dto/app_dto.go
package dto

type CreateAppRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	Platform    string `json:"platform" binding:"required"`     // "play_store" | "app_store"
	PackageName string `json:"package_name" binding:"required"` // com.example.app
	Name        string `json:"name" binding:"required"`         // human readable app name
}
