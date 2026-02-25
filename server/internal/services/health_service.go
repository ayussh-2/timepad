package services

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *HealthService) GetHealth() HealthStatus {
	return HealthStatus{
		Status:  "ok",
		Message: "Server is running",
	}
}

func (s *HealthService) Ping() string {
	return "pong"
}
