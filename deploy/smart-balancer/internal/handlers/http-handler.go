package handlers

import (
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"net/http"
	"net/http/httputil"
)

type WorkerFunc func(ctx context.Context, nodeManager NodeManager)

type NodeSelector interface {
	SelectNode(ctx context.Context, nodes []*models.Node) (*models.Node, error)
	Name() string

	PrepareData() (worker WorkerFunc)
}

type NodeManager interface {
	GetNodesForService(serviceName string) ([]*models.Node, error)
	GetAllNodesFromAllServices() ([]*models.Node, error)
}

func NewProxyHandler(strategy NodeSelector, nodeManager NodeManager) http.Handler {
	proxy := &httputil.ReverseProxy{
		Transport: http.DefaultTransport,
	}

	proxy.Director = func(req *http.Request) {
		ctx := req.Context()
		nodes, err := nodeManager.GetNodesForService("gateway")
		if err != nil {
			logger.Errorf(ctx, "Error getting nodes: %v", err)
			return
		}
		if len(nodes) == 0 {
			logger.Warnf(ctx, "No nodes available")
			return
		}

		targetNode, err := strategy.SelectNode(ctx, nodes)
		if err != nil {
			logger.Errorf(ctx, "Error selecting node: %v", err)
			return
		}

		//// --- ИЗМЕНЕНИЕ ЗДЕСЬ ---
		//// Для каждого запроса мы "подсовываем" свой Transport, который знает,
		//// на какую ноду мы отправляем запрос.
		//proxy.Transport = &LatencyTrackingTransport{
		//	BaseTransport: http.DefaultTransport,
		//	node:          targetNode, // Передаем указатель на выбранную ноду
		//}

		// Перенаправляем запрос (этот код без изменений)
		req.URL.Scheme = targetNode.URL.Scheme
		req.URL.Host = targetNode.URL.Host
		req.Host = targetNode.URL.Host
		req.RequestURI = ""
	}

	return proxy
}
