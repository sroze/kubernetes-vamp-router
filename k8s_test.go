package k8svamprouter

import (
	"errors"
	"github.com/DATA-DOG/godog/gherkin"
	"k8s.io/kubernetes/pkg/api"
)

type InMemoryServiceRepository struct {
	Services map[string]*api.Service
}

func (repository *InMemoryServiceRepository) Get(name string) (*api.Service, error) {
	service, found := repository.Services[name]
	if found {
		return service, nil
	}

	return nil, errors.New("Service do not exists")
}

func (repository *InMemoryServiceRepository) Update(service *api.Service) (*api.Service, error) {
	repository.Services[service.ObjectMeta.Name] = service

	return service, nil
}

func GetOrCreateService(repository *InMemoryServiceRepository, name string) *api.Service {
	service, err := repository.Get(name)
	if err != nil {
		service = &api.Service{
			ObjectMeta: api.ObjectMeta{
				Name: name,
			},
		}

		repository.Update(service)
	}

	return service
}

var repository *InMemoryServiceRepository

func NewInMemoryServiceRepository() *InMemoryServiceRepository {
	repository = &InMemoryServiceRepository{
		Services: make(map[string]*api.Service),
	}

	return repository
}

// FEATURES
func theKsServiceisInTheNamespace(serviceName string, namespace string) error {
	service := GetOrCreateService(repository, serviceName)
	service.ObjectMeta.Namespace = namespace

	_, err := repository.Update(service)

	return err
}

func theKsServiceIPIs(serviceName string, IP string) error {
	service := GetOrCreateService(repository, serviceName)
	service.Spec.ClusterIP = IP

	_, err := repository.Update(service)

	return err
}

func theKsServicehasTheFollowingAnnotations(serviceName string, annotationsTable *gherkin.DataTable) error {
	service := GetOrCreateService(repository, serviceName)
	annotations := make(map[string]string)

	for i, row := range annotationsTable.Rows {
		if i == 0 {
			// Skip the headers
			continue
		}

		annotations[row.Cells[0].Value] = row.Cells[1].Value
	}

	service.ObjectMeta.Annotations = annotations

	_, err := repository.Update(service)

	return err
}
