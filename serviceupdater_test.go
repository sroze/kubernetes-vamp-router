package k8svamprouter

import (
	"github.com/DATA-DOG/godog"
	"k8s.io/kubernetes/pkg/api"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	"errors"
)

type InMemoryServiceRepository struct {
}

func(repository *InMemoryServiceRepository) Update(service *api.Service) (*api.Service, error) {
	return service, nil
}

type InMemoryVampRouterClient struct  {
	Routes map[string]*vamprouter.Route
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
func aKsServiceNamedIsCreatedInTheNamespace(serviceName string, namespaceName string) error {
	return updater.CreateServiceRoute(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Name: serviceName,
			Namespace: namespaceName,
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
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		return err
	}

	return nil
}

func featureContext(s *godog.Suite) {
	s.BeforeScenario(func(interface{}) {
		updater = &ServiceUpdater{
			ServiceRepository: &InMemoryServiceRepository{},
			RouterClient: &InMemoryVampRouterClient{
				Routes: make(map[string]*vamprouter.Route),
			},
			Configuration: Configuration{
				RootDns: "example.com",
			},
		}
	})

	s.Step(`^a k8s service named "([^"]*)" is created in the namespace "([^"]*)"$`, aKsServiceNamedIsCreatedInTheNamespace)
	s.Step(`^the vamp service "([^"]*)" should be created$`, theVampServiceShouldBeCreated)
	s.Step(`^the vamp route "([^"]*)" should be created$`, theVampRouteShouldBeCreated)
	s.Step(`^a vamp route named "([^"]*)" already exists$`, aVampRouteNamedAlreadyExists)
	s.Step(`^the vamp filter named "([^"]*)" should be created$`, theVampFilterNamedShouldBeCreated)
}
