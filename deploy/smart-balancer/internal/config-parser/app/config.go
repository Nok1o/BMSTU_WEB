package app

import (
	"balancer/internal/config-parser/balancers/latency"
	smartbalancer "balancer/internal/config-parser/balancers/smart-balancer"
	"balancer/internal/config-parser/kubernetes"
	"balancer/internal/config-parser/logger"
	"os"

	"gopkg.in/yaml.v2"
)

type ApplicationConfig struct {
	ServerPort   uint   `yaml:"server_port"`
	BalancerType string `yaml:"balancer_type"`

	KubernetesCfg *kubernetes.KubernetesConfig `yaml:"kubernetes"`
	LoggerCfg     *logger.LoggerConfig         `yaml:"logger"`

	LatencyCfg       *latency.LatencyBalancerConfig     `yaml:"latency"`
	SmartBalancerCfg *smartbalancer.SmartBalancerConfig `yaml:"smart_balancer"`
}

func LoadApplicationConfig(path string) (*ApplicationConfig, error) {
	// parse yaml file into config
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ApplicationConfig

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
