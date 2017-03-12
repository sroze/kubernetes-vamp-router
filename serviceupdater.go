package k8svamprouter

import (
	api "k8s.io/client-go/pkg/api/v1"
	"log"
	"fmt"
	"strings"
)

type ServiceRepository interface {
	Update(service *api.Service) (*api.Service, error)
}

type Configuration struct {
	RootDns string
}

type ServiceUpdater struct {
	// Kubernetes client
	ServiceRepository ServiceRepository

	// Updater configuration
	Configuration Configuration
}

func ServiceHasLoadBalancerAddress(service *api.Service) bool {
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return false
	}

	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if ingress.IP != "" || ingress.Hostname != "" {
			return true
		}
	}

	return false
}

// START
// Implementation of `ObjectRoutingResolver`
//
func (su *ServiceUpdater) UpdateObjectWithDomainNames(object KubernetesBackendObject) error {
	service, ok := object.(*api.Service)
	if !ok {
		return fmt.Errorf("Get get only from `Service` objects")
	}

	if ServiceHasLoadBalancerAddress(service) {
		log.Println("The route was found and the service has an address, not updating the status")

		return nil
	}

	domainNames, err := su.GetDomainNames(service)
	if err != nil {
		return err
	}

	log.Println("Found route for the service", service.ObjectMeta.Name, "updating the service load-balancer status")
	service.Status = api.ServiceStatus{
		LoadBalancer: api.LoadBalancerStatus{
			Ingress: []api.LoadBalancerIngress{
				api.LoadBalancerIngress{
					Hostname: domainNames[0],
				},
			},
		},
	}

	_, err = su.ServiceRepository.Update(service)

	return err
}

func (su *ServiceUpdater) GetDomainNames(object KubernetesBackendObject) ([]string, error) {
	service, ok := object.(*api.Service)
	if !ok {
		return nil, fmt.Errorf("Get get only from `Service` objects")
	}

	domainNames := GetDomainNamesFromServiceAnnotations(service)

	// Add the default domain name
	domainNames = append(domainNames, strings.Join([]string{
		GetServiceRouteName(service),
		su.Configuration.RootDns,
	}, "."))

	return domainNames, nil
}

func (su *ServiceUpdater) GetRouteName(object KubernetesBackendObject) (string, error) {
	service, ok := object.(*api.Service)
	if !ok {
		return "", fmt.Errorf("Get get only from `Service` objects")
	}

	return GetServiceRouteName(service), nil
}

func (su *ServiceUpdater) GetBackendAddress(object KubernetesBackendObject) (string, error) {
	service, ok := object.(*api.Service)
	if !ok {
		return "", fmt.Errorf("Get get only from `Service` objects")
	}

	return service.Spec.ClusterIP, nil
}

// Implementation of `ObjectRoutingResolver`
// END
