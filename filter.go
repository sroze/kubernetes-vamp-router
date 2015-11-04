package main

import (
	"k8s.io/kubernetes/pkg/api"
	"log"
)

func ShouldUpdateServiceRoute(service *api.Service) bool {
	if service.Spec.Type != api.ServiceTypeLoadBalancer {
		log.Println("Skipping service", service.ObjectMeta.Name, "as it is not a LoadBalancer")

		return false
	}

	return true
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
