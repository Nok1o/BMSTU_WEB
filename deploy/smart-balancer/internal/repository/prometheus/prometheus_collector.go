package prometheus

import (
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
)

// NodeMetricsSnapshot содержит "сырые" метрики для одного узла, полученные из Prometheus.
type NodeMetricsSnapshot struct {
	IsHealthy    bool
	CPULoad      float64
	MemoryUsage  float64
	InFlightReqs float64
}

// PrometheusCollector - клиент для сбора метрик из Prometheus.
type PrometheusCollector struct {
	promAPI   v1.API
	namespace string // Неймспейс, в котором ищем метрики
}

// NewPrometheusCollector создает новый экземпляр коллектора.
func NewPrometheusCollector(promURL, namespace string) (*PrometheusCollector, error) {
	client, err := api.NewClient(api.Config{
		Address: promURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}
	return &PrometheusCollector{
		promAPI:   v1.NewAPI(client),
		namespace: namespace,
	}, nil
}

// FetchAllMetrics выполняет запросы в Prometheus и возвращает map метрик для каждого пода.
// Ключ карты - имя пода (например, "gateway-deployment-xxxx-yyyy").
func (pc *PrometheusCollector) FetchAllMetrics(ctx context.Context) (map[string]models.SmartBalancerNodeMetrics, error) {
	// Определяем PromQL-запросы для всех наших метрик
	queries := map[string]string{
		"up":         fmt.Sprintf(`up{namespace="%s", job="quickflow-app"}`, pc.namespace),
		"cpu":        fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s", container!=""}[1m])) by (pod)`, pc.namespace),
		"memory":     fmt.Sprintf(`sum(container_memory_usage_bytes{namespace="%s", container!=""}) by (pod)`, pc.namespace),
		"in_flight":  fmt.Sprintf(`sum(QuickFlow_in_flight_requests{namespace="%s"}) by (pod)`, pc.namespace),
		"goroutines": fmt.Sprintf(`sum(go_goroutines{namespace="%s", job="quickflow-app"}) by (pod)`, pc.namespace),
		// 95-й перцентиль времени ответа за последние 5 минут
		"p95_latency": fmt.Sprintf(`histogram_quantile(0.95, sum(rate(QuickFlow_Timings_bucket{namespace="%s"}[5m])) by (le, pod))`, pc.namespace),
		// Процент ошибок за последние 5 минут
		"error_rate": fmt.Sprintf(`(sum(rate(QuickFlow_Errors{namespace="%s"}[5m])) by (pod) / sum(rate(QuickFlow_Hits{namespace="%s"}[5m])) by (pod)) * 100`, pc.namespace, pc.namespace),
		// Средняя длительность пауз GC
		"gc_pause": fmt.Sprintf(`rate(go_gc_duration_seconds_sum{namespace="%s"}[1m]) / rate(go_gc_duration_seconds_count{namespace="%s"}[1m]) * 1000`, pc.namespace, pc.namespace),
	}

	// --- Параллельное выполнение запросов (этот код не меняется) ---
	resultsChan := make(chan *queryResult, len(queries))
	var wg sync.WaitGroup
	for name, q := range queries {
		wg.Add(1)
		go pc.executeQuery(ctx, name, q, &wg, resultsChan)
	}

	wg.Wait()
	close(resultsChan)

	rawMetrics := make(map[string]prommodel.Vector)
	for result := range resultsChan {
		if result.err != nil {
			logger.Errorf(ctx, "Prometheus query '%s' failed: %v", result.name, result.err)
			continue
		}
		rawMetrics[result.name] = result.vector
	}

	return pc.parseMetrics(rawMetrics), nil
}

// queryResult - вспомогательная структура для параллельных запросов.
type queryResult struct {
	name   string
	vector prommodel.Vector
	err    error
}

// executeQuery выполняет один PromQL-запрос.
func (pc *PrometheusCollector) executeQuery(ctx context.Context, name, query string, wg *sync.WaitGroup, resultsChan chan<- *queryResult) {
	defer wg.Done()
	res, warnings, err := pc.promAPI.Query(ctx, query, time.Now())
	if err != nil {
		resultsChan <- &queryResult{name: name, err: err}
		return
	}
	if len(warnings) > 0 {
		logger.Warnf(ctx, "Prometheus query '%s' warnings: %v", name, warnings)
	}

	vector, ok := res.(prommodel.Vector)
	if !ok {
		resultsChan <- &queryResult{name: name, err: fmt.Errorf("result is not a vector type")}
		return
	}
	resultsChan <- &queryResult{name: name, vector: vector}
}

// parseMetrics преобразует сырые векторы Prometheus в удобный map.
func (pc *PrometheusCollector) parseMetrics(rawMetrics map[string]prommodel.Vector) map[string]models.SmartBalancerNodeMetrics {
	snapshots := make(map[string]models.SmartBalancerNodeMetrics)

	// Функция-хелпер для безопасного добавления метрик
	upsert := func(ipAddress string, updateFunc func(s *models.SmartBalancerNodeMetrics)) {
		if ipAddress == "" {
			return
		}
		snapshot, ok := snapshots[ipAddress]
		if !ok {
			snapshot = models.SmartBalancerNodeMetrics{}
		}
		updateFunc(&snapshot)
		snapshots[ipAddress] = snapshot
	}

	// Обрабатываем все метрики, которые приходят с меткой 'instance' (IP:Port)
	// Это и `up` от service discovery, и `go_*`, и `QuickFlow_*`
	for metricName, vec := range rawMetrics {
		for _, sample := range vec {
			instance, ok := sample.Metric["instance"]
			if !ok {
				continue
			}

			// Извлекаем чистый IP из "IP:Port"
			ip, _, err := net.SplitHostPort(string(instance))
			if err != nil {
				// Если порта нет (например, от node-exporter), instance может быть просто IP
				if strings.Contains(string(instance), ":") {
					continue // Не смогли распарсить, пропускаем
				}
				ip = string(instance)
			}

			value := float64(sample.Value)

			switch metricName {
			case "up":
				upsert(ip, func(s *models.SmartBalancerNodeMetrics) { s.IsHealthy = value == 1 })
			case "cpu":
				// ВАЖНО: нужно убедиться, что метрика CPU тоже имеет метку 'instance',
				// а не только 'pod'. Если ее нет, нужно будет изменить PromQL запрос,
				// чтобы сделать join или использовать другую метрику.
				// Но обычно `kubernetes_sd_configs` добавляет ее.
				upsert(ip, func(s *models.SmartBalancerNodeMetrics) { s.CPULoad = value })
			case "memory":
				upsert(ip, func(s *models.SmartBalancerNodeMetrics) { s.MemoryUsage = value })
			case "in_flight":
				upsert(ip, func(s *models.SmartBalancerNodeMetrics) { s.InFlightRequests = int64(value) })
				// ... и так далее для p95_latency, error_rate, etc.
				// Убедитесь, что их PromQL запросы тоже группируются `by (instance)`
			}
		}
	}

	logger.Debugf(context.Background(), "snapshots: %v", snapshots)
	return snapshots
}
