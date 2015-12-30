package k8svamprouter

import (
	"k8s.io/kubernetes/pkg/api"
	"log"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
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

	// Vamp Router client
	RouterClient vamprouter.Interface

	// Updater configuration
    Configuration Configuration
}

func (su *ServiceUpdater) UpdateServiceRouting(service *api.Service) error {
	if !su.ServiceRouteIsConfigured(service) {
		err := su.CreateServiceRoute(service)
		if err != nil {
			log.Println("Unable to create service route", err)

			return err
		}
	}

	if ServiceHasLoadBalancerAddress(service) {
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

	_, err := su.ServiceRepository.Update(service)
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

func (su *ServiceUpdater) ServiceRouteIsConfigured(service *api.Service) bool {
	routeName := su.GetServiceRouteName(service)
	route, err := su.RouterClient.GetRoute("http")
	if err != nil {
		return false
	}

	for _, filter := range route.Filters {
		if filter.Name == routeName {
			return true
		}
	}

	return false
}

func (su *ServiceUpdater) CreateServiceRoute(service *api.Service) error {
	httpRoute, err := su.RouterClient.GetRoute("http")
	if err != nil {
		httpRoute, err = su.RouterClient.CreateRoute(&vamprouter.Route{
			Name: "http",
			Port: 80,
			Protocol: vamprouter.ProtocolHttp,
		})

		if err != nil {
			log.Println("Unable to create the HTTP route", err)

			return err
		}
	}

	// Create the front-end filter
	routeName := su.GetServiceRouteName(service)
	httpRoute.Filters = append(httpRoute.Filters, vamprouter.Filter{
		Name: routeName,
		Condition: "hdr(Host) -i "+su.GetDomainNameFromService(service),
		Destination: routeName,
	})

	// Create the associated service
	serviceHost := service.Spec.ClusterIP
	httpRoute.Services = append(httpRoute.Services, vamprouter.Service{
		Name: routeName,
		Weight: 0,
		Servers: []vamprouter.Server{
			vamprouter.Server{
				Name: routeName,
				Host: serviceHost,
				Port: 80,
			},
		},
	})

    _, err = su.RouterClient.UpdateRoute(httpRoute)

	return err
}
