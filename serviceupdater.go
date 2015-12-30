package k8svamprouter

import (
	"k8s.io/kubernetes/pkg/api"
	"log"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	"errors"
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

	// Vamp Router client
	RouterClient vamprouter.Interface

	// Updater configuration
    Configuration Configuration
}

func (su *ServiceUpdater) UpdateServiceRouting(service *api.Service) error {
	err := su.UpdateRouteIfNeeded(service)
	if err != nil {
		log.Println("Unable to update service route", err)

		return err
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

	_, err = su.ServiceRepository.Update(service)
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

func (su *ServiceUpdater) UpdateRouteIfNeeded(service *api.Service) error {
	routeName := su.GetServiceRouteName(service)
	route, err := su.RouterClient.GetRoute("http")
	if err != nil {
		route, err = su.RouterClient.CreateRoute(&vamprouter.Route{
			Name: "http",
			Port: 80,
			Protocol: vamprouter.ProtocolHttp,
		})

		if err != nil {
			log.Println("Unable to create the HTTP route", err)

			return err
		}
	}

	updated := false
	filter, err := GetFilterInRoute(route, routeName)
	if err != nil {
		filter = &vamprouter.Filter{
			Name: routeName,
			Condition: "hdr(Host) -i "+su.GetDomainNameFromService(service),
			Destination: routeName,
		}

		route.Filters = append(route.Filters, *filter)
		updated = true
	}

	// Create the associated service
	routeService, err := GetServiceInRoute(route, routeName)
	if err != nil {
		route.Services = append(route.Services, vamprouter.Service{
			Name: routeName,
			Weight: 0,
		})

		routeService = &route.Services[len(route.Services) - 1]
		updated = true
	}

	serviceHost := service.Spec.ClusterIP
	if len(routeService.Servers) != 1 || routeService.Servers[0].Host != serviceHost {
		routeService.Servers = []vamprouter.Server{
			vamprouter.Server{
				Name: routeName,
				Host: serviceHost,
				Port: 80,
			},
		}

		err = ReplaceServiceInRoute(route, routeName, routeService)
		if err != nil {
			return err
		}

		updated = true
	}

	if updated {
		_, err = su.RouterClient.UpdateRoute(route)

		return err
	}

	return nil
}

func GetFilterInRoute(route *vamprouter.Route, filterName string) (*vamprouter.Filter, error) {
	for _, filter := range route.Filters {
		if filter.Name == filterName {
			return &filter, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Unable to find filter named %s", filterName))
}

func ReplaceServiceInRoute(route *vamprouter.Route, serviceName string, service *vamprouter.Service) error {
	serviceIndex := -1
	for index, service := range route.Services {
		if service.Name == serviceName {
			serviceIndex = index
		}
	}

	if serviceIndex == -1 {
		return errors.New(fmt.Sprintf("Unable to find service named %s", serviceName))
	}

	route.Services[serviceIndex] = *service

	return nil
}

func GetServiceInRoute(route *vamprouter.Route, serviceName string) (*vamprouter.Service, error) {
	for _, service := range route.Services {
		if service.Name == serviceName {
			return &service, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Unable to find service named %s", serviceName))
}

func (su *ServiceUpdater) CreateServiceRoute(service *api.Service) error {
	return su.UpdateServiceRouting(service)
}
