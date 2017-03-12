package k8svamprouter

import (
	api "k8s.io/client-go/pkg/api/v1"
	"log"
	"fmt"
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

func ServiceExposesPort(service *api.Service, port int32) bool {
	for _, exposedPort := range service.Spec.Ports {
		if exposedPort.Port == port {
			return true
		}
	}

	return false
}

// START
// Implementation of `ObjectRoutingResolver`
//
func (su *ServiceUpdater) ShouldHandleObject(object KubernetesBackendObject) bool {
	service, ok := object.(*api.Service)
	if !ok {
		log.Println("[error] Get get only from `Service` objects")

		return false
	}

	if service.Spec.Type != api.ServiceTypeLoadBalancer {
		log.Println("Skipping service", service.ObjectMeta.Name, "as it is not a LoadBalancer")

		return false
	} else if !ServiceExposesPort(service, 80) {
		log.Println("Skipping service", service.ObjectMeta.Name, "because HTTP port is not exposed, other ports are NOT SUPPORTED")

		return false
	}

	return true
}

func (su *ServiceUpdater) UpdateObjectWithDomainNames(object KubernetesBackendObject, domainNames []string) error {
	service, ok := object.(*api.Service)
	if !ok {
		return fmt.Errorf("Get get only from `Service` objects")
	}

	if ServiceHasLoadBalancerAddress(service) {
		log.Println("The route was found and the service has an address, not updating the status")

		return nil
	}

	log.Println("Found route for the service", service.ObjectMeta.Name, "updating the service load-balancer status")
	service.Status = api.ServiceStatus{
		LoadBalancer: CreateLoadBalancerStatusFromDomainNames(domainNames),
	}

	_, err := su.ServiceRepository.Update(service)

	return err
}

func (su *ServiceUpdater) GetDomainNames(object KubernetesBackendObject) ([]string, error) {
	service, ok := object.(*api.Service)
	if !ok {
		return nil, fmt.Errorf("Get get only from `Service` objects")
	}

	domainNames := GetDomainNamesFromServiceAnnotations(service)

	// Add the default domain name
	domainSeparator := GetDomainSeparator()
	domainNames = append(domainNames, GetRouteNameFromObjectMetadata(service.ObjectMeta, domainSeparator)+su.Configuration.RootDns)

	return domainNames, nil
}

func (su *ServiceUpdater) GetRouteName(object KubernetesBackendObject) (string, error) {
	service, ok := object.(*api.Service)
	if !ok {
		return "", fmt.Errorf("Get get only from `Service` objects")
	}

	return GetRouteNameFromObjectMetadata(service.ObjectMeta, GetDomainSeparator()), nil
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
