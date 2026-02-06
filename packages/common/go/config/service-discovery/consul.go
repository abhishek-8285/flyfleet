package servicediscovery

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type ServiceDiscovery interface {
	RegisterService(service *ServiceRegistration) error
	DeregisterService(serviceID string) error
	GetService(serviceName string) ([]*ServiceInstance, error)
	WatchService(serviceName string, handler ServiceChangeHandler) error
	HealthCheck(serviceID string, status string) error
}

type ConsulDiscovery struct {
	client *api.Client
}

func NewConsul(config Config) (ServiceDiscovery, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address
	
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	
	return &ConsulDiscovery{client: client}, nil
}

func (cd *ConsulDiscovery) RegisterService(reg *ServiceRegistration) error {
	service := &api.AgentServiceRegistration{
		ID:      reg.ID,
		Name:    reg.Name,
		Address: reg.Address,
		Port:    reg.Port,
		Tags:    reg.Tags,
		Meta:    reg.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", reg.Address, reg.Port),
			Interval: "30s",
			Timeout:  "5s",
		},
	}
	
	return cd.client.Agent().ServiceRegister(service)
}

func (cd *ConsulDiscovery) DeregisterService(serviceID string) error {
	return cd.client.Agent().ServiceDeregister(serviceID)
}

func (cd *ConsulDiscovery) GetService(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := cd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}
	
	instances := make([]*ServiceInstance, len(services))
	for i, serviceEntry := range services {
		instances[i] = &ServiceInstance{
			ID:      serviceEntry.Service.ID,
			Name:    serviceEntry.Service.Service,
			Address: serviceEntry.Service.Address,
			Port:    serviceEntry.Service.Port,
			Tags:    serviceEntry.Service.Tags,
			Meta:    serviceEntry.Service.Meta,
		}
	}
	
	return instances, nil
}

func (cd *ConsulDiscovery) WatchService(serviceName string, handler ServiceChangeHandler) error {
	// Implementation for watching service changes
	// This would typically use Consul's blocking queries or watches
	return nil
}

func (cd *ConsulDiscovery) HealthCheck(serviceID string, status string) error {
	return cd.client.Agent().UpdateTTL("service:"+serviceID, "", status)
}

type Config struct {
	Address string
	Token   string
}

func DefaultConfig() Config {
	return Config{
		Address: "localhost:8500",
		Token:   "",
	}
}

type ServiceRegistration struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

type ServiceInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

// ServiceChangeHandler handles service instance changes
type ServiceChangeHandler func(instances []*ServiceInstance)

// Constants for health status
const (
	HealthPassing  = "passing"
	HealthWarning  = "warning"
	HealthCritical = "critical"
	HealthMaint    = "maintenance"
)