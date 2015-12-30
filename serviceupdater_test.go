package k8svamprouter

import (
	"github.com/DATA-DOG/godog"
	"k8s.io/kubernetes/pkg/api"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	"errors"
	"fmt"
)

type InMemoryVampRouterClient struct  {
	Routes map[string]*vamprouter.Route
	UpdatedRoutes []*vamprouter.Route
}

func NewInMemoryVampRouterClient() *InMemoryVampRouterClient {
	client := &InMemoryVampRouterClient{
		Routes: make(map[string]*vamprouter.Route),
	}

	client.Clear()

	return client
}

func(client *InMemoryVampRouterClient) Clear() {
	client.UpdatedRoutes = []*vamprouter.Route{}
}

func(client *InMemoryVampRouterClient) GetRoute(name string) (*vamprouter.Route, error) {
	route, found := client.Routes[name]
	if found {
		return route, nil
	}

	return nil, errors.New("Route do not exists")
}

func(client *InMemoryVampRouterClient) UpdateRoute(route *vamprouter.Route) (*vamprouter.Route, error) {
	_, found := client.Routes[route.Name]
	if !found {
		return nil, errors.New("Route not found")
	}

	client.Routes[route.Name] = route;

	return route, nil
}

func(client *InMemoryVampRouterClient) CreateRoute(route *vamprouter.Route) (*vamprouter.Route, error) {
	_, found := client.Routes[route.Name]
	if found {
		return nil, errors.New("Route already exists")
	}

	client.Routes[route.Name] = route

	return route, nil
}

var updater *ServiceUpdater

func GetCreatedServiceInRoute(route *vamprouter.Route, serviceName string) (vamprouter.Service, error) {
	for _, service := range route.Services {
		if service.Name == serviceName {
			return service, nil
		}
	}

	return vamprouter.Service{}, errors.New("Service not found")
}

func GetCreatedFilterInRoute(route *vamprouter.Route, filterName string) (vamprouter.Filter, error) {
	for _, filter := range route.Filters {
		if filter.Name == filterName {
			return filter, nil
		}
	}

	return vamprouter.Filter{}, errors.New("Filter not found")
}

/**
 * GIVEN
 */
func aVampRouteNamedAlreadyExists(routeName string) error {
	_, err := updater.RouterClient.CreateRoute(&vamprouter.Route{
		Name: "http",
		Port: 80,
		Protocol: vamprouter.ProtocolHttp,
	})

	return err
}

/**
 * WHEN
 */

func theKsServiceNamedisCreated(serviceName string) error {
	service, err := repository.Get(serviceName)
	if err != nil {
		return err
	}

	return updater.CreateServiceRoute(service)
}

func theKsServiceNamedisUpdated(serviceName string) error {
	service, err := repository.Get(serviceName)
	if err != nil {
		return err
	}

	return updater.UpdateServiceRouting(service)
}

func aKsServiceNamedIsCreatedInTheNamespace(serviceName string, namespaceName string) error {
	return updater.CreateServiceRoute(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Name: serviceName,
			Namespace: namespaceName,
		},
	})
}

func aKsServiceNamedIsCreatedInTheNamespaceWithTheIP(serviceName string, namespaceName string, IP string) error {
	return updater.CreateServiceRoute(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Name: serviceName,
			Namespace: namespaceName,
		},
		Spec: api.ServiceSpec{
			ClusterIP: IP,
		},
	})
}

func aKsServiceNamedIsUpdatedInTheNamespaceWithTheIP(serviceName string, namespaceName string, IP string) error {
	return updater.UpdateServiceRouting(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Name: serviceName,
			Namespace: namespaceName,
		},
		Spec: api.ServiceSpec{
			ClusterIP: IP,
		},
	})
}

/**
 * THEN
 */

func theVampServiceShouldBeCreated(serviceName string) error {
	route, err := updater.RouterClient.GetRoute("http")
	if err != nil {
		return err
	}

	_, err = GetCreatedServiceInRoute(route, serviceName)
	return err
}

func theVampRouteShouldBeCreated(routeName string) error {
	_, err := updater.RouterClient.GetRoute(routeName)

	return err
}

func theVampFilterNamedShouldBeCreated(filterName string) error {
	route, err := updater.RouterClient.GetRoute("http")
	if err != nil {
		return err
	}

	_, err = GetCreatedFilterInRoute(route, filterName)
	return err
}

func theVampRouteShouldNotBeUpdated() error {
	client := updater.RouterClient.(*InMemoryVampRouterClient)
	defer client.Clear()

	if len(client.UpdatedRoutes) > 0 {
		return errors.New(fmt.Sprintf("Found %d updated routes will expecting 0", len(client.UpdatedRoutes)))
	}

	return nil
}

func theVampRouteShouldBeUpdated() error {
	client := updater.RouterClient.(*InMemoryVampRouterClient)
	defer client.Clear()

	if len(client.UpdatedRoutes) != 0 {
		return errors.New("Found 0 updated routes will expecting at least one")
	}

	return nil
}

func theVampServiceShouldOnlyContainTheBackend(serviceName string, IP string) error {
	route, err := updater.RouterClient.GetRoute("http")
	if err != nil {
		return err
	}

	service, err := GetCreatedServiceInRoute(route, serviceName)
	if err != nil {
		return err
	}

	if 1 != len(service.Servers) {
		return errors.New(fmt.Sprintf("Expected to have 1 server in the service, found %d", len(service.Servers)))
	}

	if IP != service.Servers[0].Host {
		return errors.New(fmt.Sprintf("Expected to find a given IP, but found %s", service.Servers[0].Host))
	}

	return nil
}

func featureContext(s *godog.Suite) {
	s.BeforeScenario(func(interface{}) {
		updater = &ServiceUpdater{
			ServiceRepository: NewInMemoryServiceRepository(),
			RouterClient: NewInMemoryVampRouterClient(),
			Configuration: Configuration{
				RootDns: "example.com",
			},
		}
	})

	s.Step(`^a k8s service named "([^"]*)" is created in the namespace "([^"]*)"$`, aKsServiceNamedIsCreatedInTheNamespace)
	s.Step(`^a k8s service named "([^"]*)" is created in the namespace "([^"]*)" with the IP "([^"]*)"$`, aKsServiceNamedIsCreatedInTheNamespaceWithTheIP)
	s.Step(`^the vamp service "([^"]*)" should be created$`, theVampServiceShouldBeCreated)
	s.Step(`^the vamp route "([^"]*)" should be created$`, theVampRouteShouldBeCreated)
	s.Step(`^a vamp route named "([^"]*)" already exists$`, aVampRouteNamedAlreadyExists)
	s.Step(`^the vamp filter named "([^"]*)" should be created$`, theVampFilterNamedShouldBeCreated)
	s.Step(`^the vamp service "([^"]*)" should only contain the backend "([^"]*)"$`, theVampServiceShouldOnlyContainTheBackend)
	s.Step(`^a k8s service named "([^"]*)" is updated in the namespace "([^"]*)" with the IP "([^"]*)"$`, aKsServiceNamedIsUpdatedInTheNamespaceWithTheIP)
	s.Step(`^the vamp route should not be updated$`, theVampRouteShouldNotBeUpdated)
	s.Step(`^the vamp route should be updated$`, theVampRouteShouldBeUpdated)
	s.Step(`^the k8s service "([^"]*)" is in the namespace "([^"]*)"$`, theKsServiceisInTheNamespace)
	s.Step(`^the k8s service "([^"]*)" IP is "([^"]*)"$`, theKsServiceIPIs)
	s.Step(`^the k8s service named "([^"]*)" is created$`, theKsServiceNamedisCreated)
	s.Step(`^the k8s service named "([^"]*)" is updated$`, theKsServiceNamedisUpdated)
	s.Step(`^the k8s service "([^"]*)" has the following annotations:$`, theKsServicehasTheFollowingAnnotations)
}
