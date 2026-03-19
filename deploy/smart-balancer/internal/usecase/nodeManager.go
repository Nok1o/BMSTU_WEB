package usecase

import (
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

type BalancingStrategy string

const (
	RoundRobinBalancing BalancingStrategy = "round-robin"
	LatencyBalancing    BalancingStrategy = "latency"
	SmartBalancing      BalancingStrategy = "smart-balancer"
)

// NodeDiscoverer - интерфейс для обнаружения нод.
type NodeDiscoverer interface {
	DiscoverAllNodes(ctx context.Context) (map[string][]*models.Node, error)
}

// NodeManager - это реестр сервисов, который хранит и управляет состоянием всех нод.
type NodeManager struct {
	mu         sync.RWMutex
	services   map[string]map[string]*models.Node // serviceName -> {nodeURL -> *Node}
	discoverer NodeDiscoverer
	strategy   BalancingStrategy
}

func NewNodeManager(discoverer NodeDiscoverer, strategy string) (*NodeManager, error) {
	if strategy != string(RoundRobinBalancing) && strategy != string(LatencyBalancing) && strategy != string(SmartBalancing) {
		return nil, fmt.Errorf("unknown balancing strategy: %s", strategy)
	}

	return &NodeManager{
		services:   make(map[string]map[string]*models.Node),
		discoverer: discoverer,
		strategy:   BalancingStrategy(strategy),
	}, nil
}

// Manage запускает фоновые процессы: обнаружение и health-чеки.
func (nm *NodeManager) Manage(ctx context.Context) {
	// Запускаем немедленное обнаружение, чтобы не ждать первого тика
	nm.updateAllServices(ctx)

	// Запускаем тикер для периодического обнаружения
	go func() {
		ticker := time.NewTicker(15 * time.Second) // Обновляем топологию раз в 15 сек
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				nm.updateAllServices(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// updateAllServices - внутренняя функция для обновления реестра.
func (nm *NodeManager) updateAllServices(ctx context.Context) {
	discoveredServices, err := nm.discoverer.DiscoverAllNodes(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to discover nodes: %v", err)
		return
	}

	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Создаем новый map, чтобы атомарно заменить старый
	newServices := make(map[string]map[string]*models.Node)

	for serviceName, nodes := range discoveredServices {
		if _, ok := newServices[serviceName]; !ok {
			newServices[serviceName] = make(map[string]*models.Node)
		}
		for _, node := range nodes {
			urlStr := node.URL.String()
			// Если нода уже существует в старом реестре, переиспользуем ее объект
			if existingService, ok := nm.services[serviceName]; ok {
				if existingNode, ok2 := existingService[urlStr]; ok2 {
					newServices[serviceName][urlStr] = existingNode
					continue
				}
			}
			// Иначе это новая нода
			switch nm.strategy {
			case LatencyBalancing:
				node.Metrics = &models.LatencyData{}
			case SmartBalancing:
				node.Metrics = &models.SmartBalancerNodeMetrics{}
			}
			newServices[serviceName][urlStr] = node
		}
	}
	nm.services = newServices
	logger.Debugf(ctx, "All services discovered: %v", nm.services)
}

// GetNodesForService возвращает список нод для конкретного сервиса.
func (nm *NodeManager) GetNodesForService(serviceName string) ([]*models.Node, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	serviceNodes, ok := nm.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found or has no ready nodes", serviceName)
	}

	nodes := make([]*models.Node, 0, len(serviceNodes))
	for _, node := range serviceNodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (nm *NodeManager) GetAllNodesFromAllServices() ([]*models.Node, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*models.Node, 0)
	for _, serviceNodes := range nm.services {
		for _, node := range serviceNodes {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
