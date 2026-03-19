package handlers

import (
	"balancer/util/logger"
	"context"
	"fmt"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GrpcProxyHandler - это универсальный gRPC прокси-обработчик.
type GrpcProxyHandler struct {
	nodeManager NodeManager  // Тот же интерфейс, что и для HTTP
	strategy    NodeSelector // Тот же интерфейс стратегии
}

func NewGrpcProxyHandler(strategy NodeSelector, manger NodeManager) *GrpcProxyHandler {
	return &GrpcProxyHandler{
		strategy:    strategy,
		nodeManager: manger,
	}
}

// GetProxyDirector возвращает функцию-директор, которая будет выбирать бэкенд.
// Это самая важная часть.
func (h *GrpcProxyHandler) GetProxyDirector() proxy.StreamDirector {
	return func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, error) {

		logger.Infof(ctx, "[gRPC] Proxying '%s'", fullMethodName)
		// --- 1. Извлекаем метаданные из входящего запроса ---
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, nil, fmt.Errorf("missing metadata in gRPC context")
		}

		// --- 2. Находим наш заголовок X-Destination-Service ---
		destHeader := md.Get("x-destination-service")
		if len(destHeader) == 0 {
			return nil, nil, fmt.Errorf("missing 'x-destination-service' metadata header")
		}
		destinationService := destHeader[0]

		// --- 3. Используем нашу существующую логику для выбора ноды ---
		nodes, err := h.nodeManager.GetNodesForService(destinationService)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find nodes for service '%s': %w", destinationService, err)
		}

		targetNode, err := h.strategy.SelectNode(ctx, nodes)
		if err != nil {
			return nil, nil, fmt.Errorf("could not select a node for service '%s': %w", destinationService, err)
		}

		logger.Infof(ctx, "[gRPC] --> Proxying '%s' to service '%s', selected node: %s", fullMethodName, destinationService, targetNode.URL.Host)

		// --- 4. Создаем новое соединение с выбранным бэкендом ---
		// Удаляем наш служебный заголовок, чтобы он не ушел на бэкенд
		newMd := md.Copy()
		delete(newMd, "x-destination-service")

		outCtx := metadata.NewOutgoingContext(ctx, newMd)

		// Устанавливаем соединение с выбранным подом
		// Используем 'insecure' для коммуникации внутри кластера
		conn, err := grpc.DialContext(outCtx, targetNode.URL.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dial backend '%s': %w", targetNode.URL.Host, err)
		}

		return outCtx, conn, nil
	}
}
