package balancing_strategies

import (
	"balancer/internal/handlers"
	"balancer/internal/models"
	"balancer/internal/repository/prometheus"
	"balancer/util/logger"
	"context"
	"errors"
	"log"
	"math"
	"time"

	sb_config "balancer/internal/config-parser/balancers/smart-balancer"
)

// SmartBalancer - это стратегия, которая выбирает ноду на основе комплексной оценки метрик.
type SmartBalancer struct {
	config *sb_config.SmartBalancerConfig
	// Нам понадобится PrometheusCollector, чтобы "оживить" эту стратегию
	promCollector *prometheus.PrometheusCollector
}

// NewSmartBalancer создает новый экземпляр SmartBalancer.
func NewSmartBalancer(cfg *sb_config.SmartBalancerConfig, promCollector *prometheus.PrometheusCollector) *SmartBalancer {
	return &SmartBalancer{
		config:        cfg,
		promCollector: promCollector,
	}
}

// SelectNode - основной метод, который вычисляет оценку для каждого узла.
func (sb *SmartBalancer) SelectNode(ctx context.Context, nodes []*models.Node) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available nodes")
	}

	var bestNode *models.Node
	minScore := math.MaxFloat64

	// Для нормализации нам нужны максимальные значения метрик по всем узлам
	maxCPU, maxMemory, maxInFlight := getNormalizationFactors(nodes)

	for _, node := range nodes {
		// Убеждаемся, что у нас есть метрики для этого узла
		if node.Metrics == nil {
			logger.Warnf(ctx, "metrics is nil for node %v", node.URL)
			continue
		}

		metrics, ok := node.Metrics.(*models.SmartBalancerNodeMetrics)
		if !ok {
			logger.Warnf(ctx, "metrics is not of type SmartBalancerNodeMetrics for node %v", node.URL)
			continue // Пропускаем, если тип метрик не тот
		}

		// --- Базовая проверка на "живость" ---
		if !metrics.IsHealthy {
			logger.Infof(ctx, "node %v is unhealthy", node.URL)
			continue // Не рассматриваем "мертвые" узлы
		}

		// --- ПРОГНОЗИРОВАНИЕ ВРЕМЕНИ ВЫПОЛНЕНИЯ ---
		// Наш "прогноз" - это комбинация текущего Latency и количества активных запросов.
		// Чем больше запросов в очереди, тем дольше, вероятно, будет выполняться новый.
		// `avgTaskDuration` - среднее время выполнения одной задачи, его можно считать
		// из `QuickFlow_Timings_sum` / `QuickFlow_Timings_count`
		avgTaskDuration := 0.01 // Предположим, 10 мс на задачу (нужно считать реально)
		predictedLatency := metrics.Latency.Seconds() + (float64(metrics.InFlightRequests) * avgTaskDuration)

		// --- НОРМАЛИЗАЦИЯ МЕТРИК (приводим все к диапазону 0-1) ---
		normLatency := predictedLatency // Latency уже в секундах, его можно не нормализовать, если оно не сильно скачет
		normCPU := normalize(metrics.CPULoad, maxCPU)
		normMemory := normalize(metrics.MemoryUsage, maxMemory)
		normInFlight := normalize(float64(metrics.InFlightRequests), maxInFlight)

		// --- ВЫЧИСЛЕНИЕ ИТОГОВОЙ ОЦЕНКИ (SCORE) ---
		// Чем ниже оценка, тем лучше узел.
		score := (sb.config.WeightLatency * normLatency) +
			(sb.config.WeightCPU * normCPU) +
			(sb.config.WeightMemory * normMemory) +
			(sb.config.WeightInFlight * normInFlight)

		if score < minScore {
			minScore = score
			bestNode = node
		}
	}

	if bestNode == nil {
		// Если не нашли ни одного подходящего "умного" узла, возвращаем первый живой
		for _, node := range nodes {
			if m, ok := node.Metrics.(*models.SmartBalancerNodeMetrics); ok && m.IsHealthy {
				return node, nil
			}
		}
		return nil, errors.New("no healthy nodes available")
	}

	logger.Infof(ctx, "Strategy '%s' selected node: %s (score: %.4f)",
		sb.Name(), bestNode.URL.String(), minScore)

	return bestNode, nil
}

func (sb *SmartBalancer) Name() string {
	return "SmartPredictorBalancer"
}

// PrepareData для SmartBalancer возвращает воркер, который собирает метрики из Prometheus.
func (sb *SmartBalancer) PrepareData() handlers.WorkerFunc {
	return func(ctx context.Context, nodeManager handlers.NodeManager) {
		log.Println("SmartBalancer worker started: collecting metrics from Prometheus...")

		ticker := time.NewTicker(10 * time.Second) // Интервал сбора
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				allNodes, err := nodeManager.GetAllNodesFromAllServices()
				if err != nil {
					logger.Errorf(ctx, "Failed to get all nodes from all services: %v", err)
					continue
				}
				if len(allNodes) == 0 {
					logger.Warnf(ctx, "No nodes in all services")
					continue
				}

				// 1. Получаем все метрики из Prometheus
				snapshotsByIP, err := sb.promCollector.FetchAllMetrics(ctx)
				if err != nil {
					logger.Errorf(ctx, "SmartBalancer failed to fetch metrics: %v", err)
					continue
				}

				// 2. Обновляем каждую ноду в NodeManager'е
				for _, node := range allNodes {
					// Извлекаем IP из URL ноды
					nodeIP := node.URL.Hostname()

					// Ищем метрики для этого IP
					if snapshot, ok := snapshotsByIP[nodeIP]; ok {
						// Нашли метрики, обновляем объект Node
						// Предполагаем, что node.Metrics инициализирован как *models.NodeMetrics
						metrics, ok := node.Metrics.(*models.SmartBalancerNodeMetrics)
						if !ok {
							logger.Errorf(ctx, "Failed to cast metrics to SmartBalancerNodeMetrics")
							continue
						}
						*metrics = snapshot
					} else {
						// Для этого IP метрик нет, считаем узел нездоровым
						metrics, ok := node.Metrics.(*models.SmartBalancerNodeMetrics)
						if !ok {
							logger.Errorf(ctx, "Failed to cast metrics to SmartBalancerNodeMetrics")
							continue
						}
						metrics.IsHealthy = false
					}
				}

			case <-ctx.Done():
				log.Println("SmartBalancer worker stopped.")
				return
			}
		}
	}
}

// Вспомогательные функции для нормализации
func getNormalizationFactors(nodes []*models.Node) (maxCPU, maxMemory, maxInFlight float64) {
	for _, node := range nodes {
		if m, ok := node.Metrics.(*models.SmartBalancerNodeMetrics); ok {
			maxCPU = math.Max(maxCPU, m.CPULoad)
			maxMemory = math.Max(maxMemory, m.MemoryUsage)
			maxInFlight = math.Max(maxInFlight, float64(m.InFlightRequests))
		}
	}
	return
}

func normalize(value, max float64) float64 {
	if max == 0 {
		return 0
	}
	return value / max
}
