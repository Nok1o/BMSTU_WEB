// micro_addr/micro_addr.go
package micro_addr

import (
	"os"
	"strconv"
)

// Helper: get port from env or use default
func getPortFromEnv(envKey string, defaultValue int) int {
	if portStr := os.Getenv(envKey); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			return port
		}
	}
	return defaultValue
}

// Gateway
const (
	DefaultGatewayServiceAddrEnv = "GATEWAY_SERVICE_ADDR"
	DefaultGatewayServiceName    = "gateway service"
)

var DefaultGatewayServicePort = getPortFromEnv("GATEWAY_SERVICE_PORT", 8080)

// File Service
const (
	DefaultFileServiceAddrEnv = "FILE_SERVICE_ADDR"
	DefaultFileServiceName    = "file service"
)

var DefaultFileServicePort = getPortFromEnv("FILE_SERVICE_PORT", 8081)

// Post Service
const (
	DefaultPostServiceAddrEnv = "POST_SERVICE_ADDR"
	DefaultPostServiceName    = "post service"
)

var DefaultPostServicePort = getPortFromEnv("POST_SERVICE_PORT", 8082)

// User Service
const (
	DefaultUserServiceAddrEnv = "USER_SERVICE_ADDR"
	DefaultUserServiceName    = "user service"
)

var DefaultUserServicePort = getPortFromEnv("USER_SERVICE_PORT", 8083)

// Messenger Service
const (
	DefaultMessengerServiceAddrEnv = "MESSENGER_SERVICE_ADDR"
	DefaultMessengerServiceName    = "messenger service"
)

var DefaultMessengerServicePort = getPortFromEnv("MESSENGER_SERVICE_PORT", 8084)

// Feedback Service
const (
	DefaultFeedbackServiceAddrEnv = "FEEDBACK_SERVICE_ADDR"
	DefaultFeedbackServiceName    = "feedback service"
)

var DefaultFeedbackServicePort = getPortFromEnv("FEEDBACK_SERVICE_PORT", 8085)

// Friends Service
const (
	DefaultFriendsServiceAddrEnv = "FRIENDS_SERVICE_ADDR"
	DefaultFriendsServiceName    = "friends service"
)

var DefaultFriendsServicePort = getPortFromEnv("FRIENDS_SERVICE_PORT", 8086)

// Community Service
const (
	DefaultCommunityServiceAddrEnv = "COMMUNITY_SERVICE_ADDR"
	DefaultCommunityServiceName    = "community service"
)

var DefaultCommunityServicePort = getPortFromEnv("COMMUNITY_SERVICE_PORT", 8087)

// Max message size
const MaxMessageSize = 15 * 1024 * 1024
