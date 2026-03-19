package latency

type LatencyBalancerConfig struct {
	HealthEndpoint string `yaml:"health_endpoint"`
	HealthPort     uint   `yaml:"health_port"`
}
