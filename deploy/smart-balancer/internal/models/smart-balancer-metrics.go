package models

import "time"

type SmartBalancerNodeMetrics struct {
	// Базовое состояние
	IsHealthy bool          `json:"is_healthy"`
	Latency   time.Duration `json:"latency"` // Из активного health-чека

	// Уровень приложения (из ваших QuickFlow_* метрик)
	InFlightRequests int64   `json:"in_flight_requests"`
	ErrorRate        float64 `json:"error_rate_percent"`
	P95Latency       float64 `json:"p95_latency_seconds"` // 95-й перцентиль

	// Уровень пода/контейнера (из cAdvisor)
	CPULoad     float64 `json:"cpu_load_cores"`
	MemoryUsage float64 `json:"memory_usage_bytes"`

	// Уровень Go Runtime (опционально, но очень полезно)
	GoroutinesCount int64   `json:"goroutines_count"`
	GCPauseMs       float64 `json:"gc_pause_avg_ms"` // Средняя пауза GC
}
