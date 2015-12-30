package k8svamprouter

import (
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/api"
)

type KubernetesServiceRepository struct {
    Client client.Interface
}

func (repository *KubernetesServiceRepository) Update(service *api.Service) (*api.Service, error) {
	return repository.Client.Services(service.ObjectMeta.Namespace).Update(service)
}
