package dto

type HealthCheckResponse struct {
	Version     string `json:"version"`
	AppName     string `json:"app_name"`
	CurrentTime int64  `json:"current_time"`
}
