package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	service_discovery "quickflow/utils/service-discovery"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	"quickflow/metrics"
	grpc3 "quickflow/post_service/internal/delivery/grpc"
	"quickflow/post_service/internal/delivery/interceptor"
	"quickflow/post_service/internal/repository/postgres"
	"quickflow/post_service/internal/usecase"
	"quickflow/post_service/utils/validation"
	"quickflow/shared/client/file_service"
	userclient "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	"quickflow/shared/proto/post_service"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultPostServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConnFileService, err := service_discovery.NewGRPCClient(
		addr.DefaultFileServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	grpcConnUserService, err := service_discovery.NewGRPCClient(
		addr.DefaultUserServiceName,
		service_discovery.ModeFailover,
		interceptors.RequestIDClientInterceptor(),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := file_service.NewFileClient(grpcConnFileService)
	postValidator := validation.NewPostValidator()
	postRepo := postgres.NewPostgresPostRepository(db)
	postUseCase := usecase.NewPostUseCase(postRepo, fileService, postValidator)
	userUseCase := userclient.NewUserClient(grpcConnUserService)

	commentRepo := postgres.NewPostgresCommentRepository(db)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, fileService, postValidator)

	postMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultPostServicePort + 1000
		logger.Info(context.Background(), "Metrics server is running on :%d/metrics", metricsPort)
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDServerInterceptor(),
			interceptors.MetricsInterceptor(addr.DefaultPostServiceName, postMetrics),
			interceptor.ErrorInterceptor,
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize))
	proto.RegisterPostServiceServer(server, grpc3.NewPostServiceServer(postUseCase, userUseCase))
	proto.RegisterCommentServiceServer(server, grpc3.NewCommentServiceServer(commentUseCase, userUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
