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
	DefaultGatewayServiceName = "gateway"
)

// File Service
const (
	DefaultFileServiceName = "file-service"
)

var DefaultFileServicePort = getPortFromEnv("FILE_SERVICE_PORT", 8081)

// Post Service
const (
	DefaultPostServiceName = "post-service"
)

var DefaultPostServicePort = getPortFromEnv("POST_SERVICE_PORT", 8082)

// User Service
const (
	DefaultUserServiceName = "user-service"
)

var DefaultUserServicePort = getPortFromEnv("USER_SERVICE_PORT", 8083)

// Messenger Service
const (
	DefaultMessengerServiceName = "messenger-service"
)

var DefaultMessengerServicePort = getPortFromEnv("MESSENGER_SERVICE_PORT", 8084)

// Feedback Service
const (
	DefaultFeedbackServiceName = "feedback-service"
)

var DefaultFeedbackServicePort = getPortFromEnv("FEEDBACK_SERVICE_PORT", 8085)

// Friends Service
const (
	DefaultFriendsServiceName = "friends-service"
)

var DefaultFriendsServicePort = getPortFromEnv("FRIENDS_SERVICE_PORT", 8086)

// Community Service
const (
	DefaultCommunityServiceName = "community-service"
)

var DefaultCommunityServicePort = getPortFromEnv("COMMUNITY_SERVICE_PORT", 8087)

// Max message size
const MaxMessageSize = 15 * 1024 * 1024

var DefaultServicePorts = map[string]string{
	"gateway":           "8080",
	"file-service":      "8081",
	"post-service":      "8082",
	"user-service":      "8083",
	"messenger-service": "8084",
	"feedback-service":  "8085",
	"friends-service":   "8086",
	"community-service": "8087",
}
