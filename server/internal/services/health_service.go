package services

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type HealthService struct {
	rdb *redis.Client
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func NewHealthServiceWithRedis(rdb *redis.Client) *HealthService {
	return &HealthService{rdb: rdb}
}

type HealthStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	QueueMode string `json:"queue_mode"` // "async" | "sync"
	Redis     string `json:"redis"`      // "connected" | "unavailable"
}

func (s *HealthService) GetHealth() HealthStatus {
	redisStatus := "unavailable"
	queueMode := "sync"

	if s.rdb != nil {
		if err := s.rdb.Ping(context.Background()).Err(); err == nil {
			redisStatus = "connected"
			queueMode = "async"
		}
	}

	return HealthStatus{
		Status:    "ok",
		Message:   "Server is running",
		QueueMode: queueMode,
		Redis:     redisStatus,
	}
}

func (s *HealthService) Ping() string {
	return "pong"
}
