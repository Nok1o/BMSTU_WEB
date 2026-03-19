package internal

import (
	"balancer/internal/config-parser/app"
	"balancer/internal/handlers"
	prometheus2 "balancer/internal/repository/prometheus"
	"balancer/internal/usecase"
	"balancer/internal/usecase/balancing-strategies"
	"balancer/pkg/kubernetes"
	"balancer/util/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mwitkow/grpc-proxy/proxy"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func Run(config *app.ApplicationConfig) {
	if config == nil {
		log.Fatalf("Config is nil")
	}

	if config.LoggerCfg == nil {
		log.Fatalf("Logger config is nil")
	}

	logger.InitLogger(config.LoggerCfg.Level, os.Stdout)

	discoverer, err := kubernetes.NewDiscoverer(config.KubernetesCfg)
	if err != nil {
		log.Fatalf("Failed to create discoverer: %v", err)
	}

	// Создаем пул, который будет хранить наши ноды
	newNodeManager, err := usecase.NewNodeManager(discoverer, config.BalancerType)
	if err != nil {
		log.Fatalf("Failed to create node manager: %v", err)
	}
	go newNodeManager.Manage(context.Background())

	var currentStrategy handlers.NodeSelector
	switch config.BalancerType {
	case "latency":
		if config.LatencyCfg == nil {
			log.Fatalf("Latency balancer config is nil")
		}
		currentStrategy = balancing_strategies.NewLatencyBalancer(config.LatencyCfg.HealthEndpoint, "")
	case "round-robin":
		currentStrategy = balancing_strategies.NewRoundRobinBalancer()
	case "smart-balancer":
		if config.SmartBalancerCfg == nil {
			log.Fatalf("Smart-Balancer config is nil")
		}
		promCollector, err := prometheus2.NewPrometheusCollector(
			"http://prometheus-service.monitoring:9090",
			"default",
		)
		if err != nil {
			log.Fatalf("Failed to create prometheus collector: %v", err)
		}
		currentStrategy = balancing_strategies.NewSmartBalancer(config.SmartBalancerCfg, promCollector)
	default:
		log.Fatalf("Unknown balancer type: %s", config.BalancerType)
	}

	if worker := currentStrategy.PrepareData(); worker != nil {
		// Если он нужен, запускаем его в отдельной горутине
		logger.Infof(context.Background(), "Strategy '%s' requires data preparation. Starting worker...", currentStrategy.Name())
		go worker(context.Background(), newNodeManager)
	}

	httpProxyHandler := handlers.NewProxyHandler(currentStrategy, newNodeManager)

	//log.Printf("Smart-balancer started with '%s' strategy on :%d", currentStrategy.Name(), config.ServerPort)
	//if err = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), httpProxyHandler); err != nil {
	//	log.Fatalf("Failed to start server: %v", err)
	//}

	grpcProxyHandler := handlers.NewGrpcProxyHandler(currentStrategy, newNodeManager)

	// Создаем gRPC сервер с "прозрачным" кодеком, который пробрасывает все вызовы
	grpcServer := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()), // Это позволяет проксировать любой gRPC-сервис
		grpc.UnknownServiceHandler(proxy.TransparentHandler(grpcProxyHandler.GetProxyDirector())),
	)

	universalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, является ли это gRPC-запросом
		logger.Infof(r.Context(), "Request received: %s %s", r.Method, r.URL.Path)
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			// Если да, передаем его gRPC-серверу
			logger.Infof(r.Context(), "gRPC request received")
			grpcServer.ServeHTTP(w, r)
		} else {
			// Иначе, передаем его HTTP-прокси
			logger.Infof(r.Context(), "HTTP request received")
			httpProxyHandler.ServeHTTP(w, r)
		}
	})

	// --- 4. Запускаем один сервер с этим универсальным обработчиком ---
	serverAddr := fmt.Sprintf(":%d", config.ServerPort)
	logger.Infof(context.Background(), "Smart-balancer (HTTP & gRPC) started on %s with '%s' strategy", serverAddr, currentStrategy.Name())

	// Создаем h2c-обработчик, который оборачивает наш универсальный обработчик
	h2cHandler := h2c.NewHandler(universalHandler, &http2.Server{})

	// Запускаем сервер с этим новым обработчиком
	if err := http.ListenAndServe(serverAddr, h2cHandler); err != nil {
		log.Fatalf("Failed to start universal server: %v", err)
	}
}
