package k8svamprouter

import (
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	"log"
)

type KubernetesBackendObject interface {

}

type VampRouteManager struct {
	// Vamp Router client
	RouterClient vamprouter.Interface

	// Object Routing Resolver
	ObjectRoutingResolver ObjectRoutingResolver
}

type ObjectRoutingResolver interface {
	GetDomainNames(object KubernetesBackendObject) ([]string, error)
	GetRouteName(object KubernetesBackendObject) (string, error)
	GetBackendAddress(object KubernetesBackendObject) (string, error)
	UpdateObjectWithDomainNames(object KubernetesBackendObject, domainNames []string) error
	ShouldHandleObject(object KubernetesBackendObject) bool
}

func (rm *VampRouteManager) UpdateObjectRouting(object KubernetesBackendObject) error {
	domainNames, err := rm.UpdateRouteIfNeeded(object)
	if err != nil {
		log.Println("Unable to update object route", err)

		return err
	}

	err = rm.ObjectRoutingResolver.UpdateObjectWithDomainNames(object, domainNames)
	if err != nil {
		log.Println("Error while updating the object:", err)

		return err
	}

	log.Println("Successfully updated the object status")

	return nil
}

func (rm *VampRouteManager) RemoveObjectRouting(object KubernetesBackendObject) {
	log.Println("[TODO] Should remove service routing")
}

func (rm *VampRouteManager) CreateObjectRoute(object KubernetesBackendObject) error {
	return rm.UpdateObjectRouting(object)
}

func (rm *VampRouteManager) ShouldHandleObject(object KubernetesBackendObject) bool {
	return rm.ObjectRoutingResolver.ShouldHandleObject(object)
}

func (rm *VampRouteManager) UpdateRouteIfNeeded(object KubernetesBackendObject) ([]string, error) {
	route, err := rm.GetOrCreateHttpRoute()
	if err != nil {
		return nil, err
	}

	routeName, err := rm.ObjectRoutingResolver.GetRouteName(object)
	if err != nil {
		return nil, err
	}

	backendAddress, err := rm.ObjectRoutingResolver.GetBackendAddress(object)
	if err != nil {
		return nil, err
	}

	backend, updated, err := rm.GetCreateOrUpdateBackend(
		route,
		routeName,
		backendAddress,
	)

	if err != nil {
		return nil, err
	}

	// Create the filters
	domainNames, err := rm.ObjectRoutingResolver.GetDomainNames(object)
	if err != nil {
		return nil, err
	}

	for _, domainName := range domainNames {
		filterName := GetDNSIdentifier(domainName)
		filter, err := GetFilterInRoute(route, filterName)

		if err == nil {
			// Filter already exists, just pass
			continue
		}

		log.Println("Added the filter", filterName, "for the hostname", domainName, "to the backend", backendAddress)
		filter = &vamprouter.Filter{
			Name:        filterName,
			Condition:   "hdr(Host) -i " + domainName,
			Destination: backend.Name,
		}

		route.Filters = append(route.Filters, *filter)
		updated = true
	}

	if updated {
		_, err = rm.RouterClient.UpdateRoute(route)

		return domainNames, err
	}

	return domainNames, nil
}

func (rm *VampRouteManager) GetCreateOrUpdateBackend(route *vamprouter.Route, routeName string, backendAddress string) (*vamprouter.Service, bool, error) {
	updated := false

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
	if len(routeService.Servers) != 1 || routeService.Servers[0].Host != backendAddress {
		routeService.Servers = []vamprouter.Server{
			vamprouter.Server{
				Name: routeName,
				Host: backendAddress,
				Port: 80,
			},
		}

		err = ReplaceServiceInRoute(route, routeName, routeService)
		updated = true
	}

	return routeService, updated, err
}

func (rm *VampRouteManager) GetOrCreateHttpRoute() (*vamprouter.Route, error) {
	route, err := rm.RouterClient.GetRoute("http")
	if err != nil {
		route, err = rm.RouterClient.CreateRoute(&vamprouter.Route{
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
