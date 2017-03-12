package k8svamprouter

import (
	api "k8s.io/client-go/pkg/api/v1"
	"log"
)

func ShouldUpdateServiceRoute(service *api.Service) bool {
	if service.Spec.Type != api.ServiceTypeLoadBalancer {
		log.Println("Skipping service", service.ObjectMeta.Name, "as it is not a LoadBalancer")

		return false
	} else if !ServiceExposesPort(service, 80) {
		log.Println("Skipping service", service.ObjectMeta.Name, "because HTTP port is not exposed, other ports are NOT SUPPORTED")

		return false
	}

	return true
}

func ServiceExposesPort(service *api.Service, port int32) bool {
	for _, exposedPort := range service.Spec.Ports {
		if exposedPort.Port == port {
			return true
		}
	}

	return false
}
