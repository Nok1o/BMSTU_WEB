package kubernetes

import (
	kubernetes2 "balancer/internal/config-parser/kubernetes"
	"balancer/internal/models"
	"balancer/util/logger"
	"context"
	"fmt"
	"log"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Discoverer struct {
	clientset *kubernetes.Clientset
	cfg       *kubernetes2.KubernetesConfig
}

// NewDiscoverer создает новый экземпляр Discoverer.
func NewDiscoverer(cfg *kubernetes2.KubernetesConfig) (*Discoverer, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Discoverer{
		clientset: clientset,
		cfg:       cfg,
	}, nil
}

// DiscoverAllServices находит все сервисы в указанных неймспейсах,
// которые помечены для балансировки, и возвращает map с их нодами.
// Возвращает: map[serviceName][]*models.Node
func (d *Discoverer) DiscoverAllNodes(ctx context.Context) (map[string][]*models.Node, error) {
	// Ищем все EndpointSlice в нашем неймспейсе
	endpointSlices, err := d.clientset.DiscoveryV1().EndpointSlices(d.cfg.ServiceDiscoveryNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoint slices: %w", err)
	}

	// map для хранения результатов: serviceName -> список нод
	servicesNodes := make(map[string][]*models.Node)

	for _, slice := range endpointSlices.Items {
		// Получаем имя родительского сервиса из метки
		serviceName, ok := slice.Labels["kubernetes.io/service-name"]
		if !ok {
			continue // Пропускаем слайсы без этой метки
		}

		// (Опционально) Можно добавить фильтр, чтобы обнаруживать только те сервисы,
		// у которых есть специальная аннотация, например 'smart-balancer.io/enabled: "true"'

		for _, endpoint := range slice.Endpoints {
			if endpoint.Conditions.Ready == nil || !*endpoint.Conditions.Ready {
				continue
			}

			var protocol models.Protocol
			for _, ip := range endpoint.Addresses {
				// ВАЖНО: Нам нужны оба порта - основной и для health-check'ов
				var servicePort, healthPort int32

				for _, port := range slice.Ports {
					if port.Name != nil && port.Port != nil {
						switch *port.Name {
						case "http":
							protocol = models.HTTP
							servicePort = *port.Port
							logger.Infof(ctx, "discovered http service %s:%d", serviceName, servicePort)
						case "grpc":
							protocol = models.GRPC
							servicePort = *port.Port
							logger.Infof(ctx, "Discovered GRPC service: %s:%d", serviceName, servicePort)
						case "health", "metrics": // Имена для служебного трафика
							healthPort = *port.Port
						}
					}
				}

				// Добавляем ноду, только если нашли оба порта
				if servicePort == 0 || healthPort == 0 {
					continue
				}

				serviceURL := fmt.Sprintf("http://%s:%d", ip, servicePort)
				parsedServiceURL, err := url.Parse(serviceURL)
				if err != nil {
					continue
				}

				// Создаем ноду с двумя URL
				node := &models.Node{
					URL:      *parsedServiceURL,
					Protocol: protocol,
				}

				servicesNodes[serviceName] = append(servicesNodes[serviceName], node)
			}
		}
	}

	log.Printf("Discovery cycle finished. Found %d services.", len(servicesNodes))
	return servicesNodes, nil
}
