package main

import (
	client "github.com/kubernetes/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/api"
	"log"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
)

type Configuration struct {
	RootDns string
}

type ServiceUpdater struct {
	// Kubernetes client
	ClusterClient *client.Client

	// Vamp Router client
	RouterClient *vamprouter.Client

	// Updater configuration
    Configuration Configuration
}

func (su *ServiceUpdater) UpdateServiceRouting(service *api.Service) error {
	_, found := su.GetServiceRoute(service)
	if !found {
		_, err := su.CreateRoute(service)
		if err != nil {
			return err
		}
	} else if ServiceHasLoadBalancerAddress(service) {
		log.Println("The route was found and the service has an address, not updating the status")

		return nil
	}

	log.Println("Found route for the service", service.ObjectMeta.Name, "updating the service load-balancer status")
	service.Status = api.ServiceStatus{
		LoadBalancer: api.LoadBalancerStatus{
			Ingress: []api.LoadBalancerIngress{
				api.LoadBalancerIngress{
					Hostname: su.GetDomainNameFromService(service),
				},
			},
		},
	}

	_, err := su.ClusterClient.Services(service.ObjectMeta.Namespace).Update(service)
	if err != nil {
		log.Println("Error while updating the service:", err)

		return err
	}

	log.Println("Successfully updated the service status")

	return nil
}

func (su *ServiceUpdater) RemoveServiceRouting(service *api.Service) {
	log.Println("Should remove service routing")
}

func (su *ServiceUpdater) GetServiceRoute(service *api.Service) (*vamprouter.Route, bool) {
	routeName := su.GetServiceRouteName(service)
	route, err := su.RouterClient.GetRoute(routeName)
	if err != nil {
		return nil, false
	}

	return route, true
}

func (su *ServiceUpdater) CreateRoute(service *api.Service) (*vamprouter.Route, error) {
	routeName := su.GetServiceRouteName(service)
	serviceHost := service.Spec.ClusterIP
	route := vamprouter.Route{
		Name: routeName,
		Port: 80,
		Protocol: vamprouter.ProtocolHttp,
		Filters: []vamprouter.Filter{
			vamprouter.Filter{
				Name: "service",
				Condition: "hdr_sub(Host) "+su.GetDomainNameFromService(service),
				Destination: "service",
			},
		},
		HttpQuota: vamprouter.Quota{
			SampleWindow: "1s",
			Rate: 1000,
			ExpiryTime: "15s",
		},
		Services: []vamprouter.Service{
			vamprouter.Service{
				Name: "none",
				Weight: 100,
			},
			vamprouter.Service{
				Name: "service",
				Weight: 0,
				Servers: []vamprouter.Server{
					vamprouter.Server{
						Name: routeName,
						Host: serviceHost,
						Port: 80,
					},
				},
			},
		},
	}

    return su.RouterClient.CreateRoute(&route)
}
