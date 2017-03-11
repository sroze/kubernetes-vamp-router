package k8svamprouter

import (
	"errors"
	"fmt"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	"k8s.io/kubernetes/pkg/api"
	"log"
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
					Hostname: su.GetDomainNamesFromService(service)[0],
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
	route, err := su.GetOrCreateHttpRoute()
	if err != nil {
		return err
	}

	backend, updated, err := su.GetCreateOrUpdateBackend(route, service)
	if err != nil {
		return err
	}

	// Create the filters
	domainNames := su.GetDomainNamesFromService(service)

	for _, domainName := range domainNames {
		filterName := GetDNSIdentifier(domainName)
		filter, err := GetFilterInRoute(route, filterName)

		if err == nil {
			// Filter already exists, just pass
			continue
		}

		filter = &vamprouter.Filter{
			Name:        filterName,
			Condition:   "hdr(Host) -i " + domainName,
			Destination: backend.Name,
		}

		route.Filters = append(route.Filters, *filter)
		updated = true
	}

	if updated {
		_, err = su.RouterClient.UpdateRoute(route)

		return err
	}

	return nil
}

func (su *ServiceUpdater) GetCreateOrUpdateBackend(route *vamprouter.Route, service *api.Service) (*vamprouter.Service, bool, error) {
	updated := false
	routeName := GetServiceRouteName(service)

	// Create the backend service if it do not exists
	routeService, err := GetServiceInRoute(route, routeName)
	if err != nil {
		route.Services = append(route.Services, vamprouter.Service{
			Name:   routeName,
			Weight: 0,
		})

		routeService = &route.Services[len(route.Services)-1]
		updated = true
		err = nil
	}

	// Updates the backend if needed
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
		updated = true
	}

	return routeService, updated, err
}

func (su *ServiceUpdater) GetOrCreateHttpRoute() (*vamprouter.Route, error) {
	route, err := su.RouterClient.GetRoute("http")
	if err != nil {
		route, err = su.RouterClient.CreateRoute(&vamprouter.Route{
			Name:     "http",
			Port:     80,
			Protocol: vamprouter.ProtocolHttp,
		})

		if err != nil {
			log.Println("Unable to create the HTTP route", err)

			return nil, err
		}
	}

	return route, err
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
