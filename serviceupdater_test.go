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

func aKsServiceNamedIsCreatedInTheNamespace(serviceName string, namespaceName string) error {
	return updater.CreateServiceRoute(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Name: serviceName,
			Namespace: namespaceName,
		},
	})
}

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

func GetCreatedServiceInRoute(route *vamprouter.Route, serviceName string) (vamprouter.Service, error) {
	for _, service := range route.Services {
		if service.Name == serviceName {
			return service, nil
		}
	}

	return vamprouter.Service{}, errors.New("Service not found")
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
}
