package main

import (
	"k8s.io/kubernetes/pkg/api"
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

func ServiceExposesPort(service *api.Service, port int) bool {
	for _, exposedPort := range service.Spec.Ports {
		if exposedPort.Port == port {
			return true
		}
	}

	return false
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
