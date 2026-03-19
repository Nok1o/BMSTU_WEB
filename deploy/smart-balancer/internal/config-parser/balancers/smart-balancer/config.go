package smart_balancer

type SmartBalancerConfig struct {
	WeightLatency  float64 `yaml:"weight_latency"`
	WeightCPU      float64 `yaml:"weight_cpu"`
	WeightMemory   float64 `yaml:"weight_memory"`
	WeightInFlight float64 `yaml:"weight_in_flight"`
	// Можно добавить и другие веса
}
