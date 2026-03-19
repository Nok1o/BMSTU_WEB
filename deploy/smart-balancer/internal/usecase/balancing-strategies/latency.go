package balancing_strategies

import (
	"balancer/internal/handlers"
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// LatencyBalancer - это стратегия, которая выбирает ноду
// с наименьшим временем ответа (latency).
type LatencyBalancer struct {
	httpHealthPath string
	grpcHealthSvc  string
	checkInterval  time.Duration
	checkTimeout   time.Duration
}

func NewLatencyBalancer(httpPath, grpcService string) *LatencyBalancer {
	return &LatencyBalancer{
		httpHealthPath: httpPath,
		grpcHealthSvc:  grpcService,
		checkInterval:  5 * time.Second,
		checkTimeout:   2 * time.Second,
	}
}

// SelectNode реализует интерфейс Strategy.
// Он ищет ноду с минимальным значением Latency.
func (lb *LatencyBalancer) SelectNode(ctx context.Context, nodes []*models.Node) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available nodes")
	}

	var bestNode *models.Node = nil
	var minLatency int64 = math.MaxInt64 // Начинаем с "бесконечности"

	// Проходим по всем доступным нодам
	for _, node := range nodes {
		if node.Metrics == nil {
			node.Metrics = &models.LatencyData{}
		}

		latencyMetrics, ok := node.Metrics.(*models.LatencyData)
		if !ok {
			logger.Errorf(ctx, "could not cast node.Metrics to *models.LatencyData")
			continue
		}

		latency := latencyMetrics.GetLatency()

		// Если у ноды еще нет измерений (Latency == 0),
		// она - идеальный кандидат для первого запроса.
		if latency == 0 {
			return node, nil
		}

		// Ищем ноду с наименьшей задержкой
		if latency < minLatency {
			minLatency = latency
			bestNode = node
		}
	}

	// Если по какой-то причине не нашли лучшую ноду (маловероятно),
	// просто возвращаем первую, чтобы избежать паники.
	if bestNode == nil {
		return nodes[0], nil
	}

	logger.Infof(ctx, "Strategy '%s' selected node: %s (current latency: %.2fms)",
		lb.Name(), bestNode.URL.String(), minLatency)

	return bestNode, nil
}

// Name возвращает имя стратегии.
func (lb *LatencyBalancer) Name() string {
	return "LatencyBalancer"
}

// PrepareData для LatencyBalancer возвращает функцию-воркер для health-чеков.
func (lb *LatencyBalancer) PrepareData() handlers.WorkerFunc {
	// Создаем функцию, которая будет выполняться в фоне
	return func(ctx context.Context, nodeManager handlers.NodeManager) {
		log.Println("LatencyBalancer worker started: running health checks...")

		ticker := time.NewTicker(lb.checkInterval)
		defer ticker.Stop()

		// Запускаем первую проверку немедленно
		lb.runChecks(nodeManager)

		for {
			select {
			case <-ticker.C:
				lb.runChecks(nodeManager)
			case <-ctx.Done():
				log.Println("LatencyBalancer worker stopped.")
				return
			}
		}
	}
}

// runChecks - вспомогательная функция, которая делает одну итерацию проверок.
// Она вынесена, чтобы не дублировать код.
func (lb *LatencyBalancer) runChecks(nodeManager handlers.NodeManager) {
	nodes, err := nodeManager.GetAllNodesFromAllServices()
	if err != nil {
		logger.Errorf(context.Background(), "Error getting nodes: %v", err)
		return
	}

	for _, node := range nodes {

		switch node.Protocol {
		case models.HTTP:
			go lb.checkHTTPHealth(node)
		case models.GRPC:
			go lb.checkGRPCHealth(node)
		default:
			logger.Errorf(context.Background(), "Unknown protocol: %s", node.Protocol)
		}
	}
}

func (lb *LatencyBalancer) checkHTTPHealth(node *models.Node) {
	client := &http.Client{Timeout: lb.checkTimeout}
	healthURL := fmt.Sprintf("http://%s%s", node.URL.Host, lb.httpHealthPath)

	start := time.Now()
	resp, err := client.Get(healthURL)
	duration := time.Since(start)

	latencyMetrics, ok := node.Metrics.(*models.LatencyData)
	if !ok {
		logger.Errorf(context.TODO(), "Error casting node.Metrics to *models.LatencyData")
		return
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Errorf(context.Background(), "[HTTP] Health check FAILED for %s: %v", node.URL.String(), err)
		latencyMetrics.UpdateLatency((10 * time.Second).Milliseconds()) // Штраф
		return
	}
	defer resp.Body.Close()
	latencyMetrics.UpdateLatency(duration.Milliseconds())
}

func (lb *LatencyBalancer) checkGRPCHealth(node *models.Node) {
	// Для gRPC health-чек идет на "боевой" порт, поэтому используем ServiceURL.Host
	ctx, cancel := context.WithTimeout(context.Background(), lb.checkTimeout)
	defer cancel()

	start := time.Now()
	conn, err := grpc.DialContext(ctx, node.URL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	duration := time.Since(start)

	latencyMetrics, ok := node.Metrics.(*models.LatencyData)
	if !ok {
		logger.Errorf(context.TODO(), "Error casting node.Metrics to *models.LatencyData")
		return
	}

	if err != nil {
		logger.Errorf(context.TODO(), "[gRPC] Health check FAILED (dial) for %s: %v", node.URL, err)
		latencyMetrics.UpdateLatency(10 * time.Second.Milliseconds()) // Штраф
		return
	}
	defer conn.Close()

	healthClient := grpc_health_v1.NewHealthClient(conn)

	// Делаем сам Check-вызов
	checkCtx, checkCancel := context.WithTimeout(context.Background(), lb.checkTimeout-duration) // Оставшееся время
	defer checkCancel()

	resp, err := healthClient.Check(checkCtx, &grpc_health_v1.HealthCheckRequest{Service: lb.grpcHealthSvc})

	totalDuration := time.Since(start) // Общее время на коннект и проверку

	if err != nil || resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		logger.Errorf(ctx, "[gRPC] Health check FAILED (check) for %s: status %v, err: %v", node.URL, resp.GetStatus(), err)
		latencyMetrics.UpdateLatency(10 * time.Second.Milliseconds()) // Штраф
		return
	}

	latencyMetrics.UpdateLatency(totalDuration.Milliseconds())
}

// checkNodeHealth - хелпер для одной проверки.
func checkNodeHealth(client *http.Client, node *models.Node, healthCheckPath string) {
	start := time.Now()
	resp, err := client.Get(node.URL.String() + healthCheckPath)
	duration := time.Since(start)

	var nodeMetrics *models.LatencyData
	if node.Metrics == nil {
		nodeMetrics = &models.LatencyData{}
		node.Metrics = nodeMetrics
	} else {
		var ok bool

		nodeMetrics, ok = node.Metrics.(*models.LatencyData)
		if !ok {
			logger.Errorf(context.TODO(), "Error casting node.Metrics to *models.LatencyData")
			return
		}
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Errorf(context.TODO(), "Health check FAILED for node %s: %v", node.URL.String(), err)
		nodeMetrics.UpdateLatency(10 * time.Second.Milliseconds()) // Штрафное время
		return
	}
	defer resp.Body.Close()
	nodeMetrics.UpdateLatency(duration.Milliseconds())
}
