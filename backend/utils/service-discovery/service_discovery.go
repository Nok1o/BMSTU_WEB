package service_discovery

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	addr "quickflow/config/micro-addr"
	"quickflow/shared/logger"
	getEnv "quickflow/utils/get-env"
	"strings"
	"time"
)

type ConnectionMode string

const (
	ModeRoundRobin ConnectionMode = "round_robin"
	ModeFailover   ConnectionMode = "failover"
)

// --- STATIC RESOLVER (для round_robin) ---

type staticResolverBuilder struct {
	serviceName string
	addrs       []string
}

func (b *staticResolverBuilder) Scheme() string {
	return "static"
}

func (b *staticResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var resolved []resolver.Address
	for _, a := range b.addrs {
		resolved = append(resolved, resolver.Address{Addr: a})
	}

	cc.UpdateState(resolver.State{Addresses: resolved})
	return &nopResolver{}, nil
}

type nopResolver struct{}

func (*nopResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*nopResolver) Close()                                  {}

// --- DISCOVERY LOGIC ---

// GetServiceEndpoints возвращает все endpoints для сервиса
func GetServiceEndpoints(serviceName string) []string {
	fileServiceNames := getEnv.GetEnv("AVAILABLE_"+strings.ToUpper(serviceName)+"_NAME", serviceName)
	splittedNames := strings.Split(fileServiceNames, ",")

	fileServicePorts := getEnv.GetEnv("AVAILABLE_"+strings.ToUpper(serviceName)+"_PORT", addr.DefaultServicePorts[serviceName])
	splittedPorts := strings.Split(fileServicePorts, ",")

	var endpoints []string
	n := min(len(splittedNames), len(splittedPorts))
	for i := 0; i < n; i++ {
		endpoints = append(endpoints, splittedNames[i]+":"+splittedPorts[i])
	}

	logger.Info(context.Background(), "for service %s found these endpoints: %v", serviceName, endpoints)
	return endpoints
}

// --- ROUND ROBIN CLIENT ---

func newRoundRobinClient(serviceName string, endpoints []string, interceptors ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	builder := &staticResolverBuilder{
		serviceName: serviceName,
		addrs:       endpoints,
	}
	resolver.Register(builder)

	target := fmt.Sprintf("static:///%s", serviceName)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	}

	for _, interceptor := range interceptors {
		opts = append(opts, grpc.WithUnaryInterceptor(interceptor))
	}

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with round_robin to %s: %w", serviceName, err)
	}
	return conn, nil
}

// --- FAILOVER CLIENT ---

func newFailoverClient(serviceName string, endpoints []string, interceptors ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	}

	for _, interceptor := range interceptors {
		opts = append(opts, grpc.WithUnaryInterceptor(interceptor))
	}

	var lastErr error
	for _, endpoint := range endpoints {
		logger.Info(context.Background(), "[failover] trying to connect to %s", endpoint)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := grpc.DialContext(ctx, endpoint, opts...)
		cancel()
		if err == nil {
			logger.Info(context.Background(), "[failover] connected to %s", endpoint)
			return conn, nil
		}
		logger.Warn(context.Background(), "[failover] failed to connect to %s: %v", endpoint, err)
		lastErr = err
	}

	return nil, fmt.Errorf("all endpoints failed for %s: %v", serviceName, lastErr)
}

// --- PUBLIC ENTRY POINT ---

// NewGRPCClient создает клиент с балансировкой или failover, в зависимости от режима
func NewGRPCClient(serviceName string, mode ConnectionMode, interceptors ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	endpoints := GetServiceEndpoints(serviceName)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints found for service: %s", serviceName)
	}

	switch mode {
	case ModeRoundRobin:
		return newRoundRobinClient(serviceName, endpoints, interceptors...)
	case ModeFailover:
		return newFailoverClient(serviceName, endpoints, interceptors...)
	default:
		return nil, fmt.Errorf("unknown connection mode: %s", mode)
	}
}

// --- Вспомогательная функция ---

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
