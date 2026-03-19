package models

import "sync"

type LatencyData struct {
	mu      sync.RWMutex
	latency int64
}

// UpdateLatency обновляет время ответа ноды, используя экспоненциальное сглаживание.
// Alfa (α) - коэффициент сглаживания (0.1 - 0.3 - хорошее значение).
// Чем выше alfa, тем больше вес у последнего измерения.
func (d *LatencyData) UpdateLatency(lastLatency int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	const alfa = 0.2

	if d.latency == 0 {
		d.latency = lastLatency
	} else {
		d.latency = int64((alfa * float64(lastLatency)) + (1-alfa)*float64(d.latency))
	}
}

func (d *LatencyData) GetLatency() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.latency
}
