package k8svamprouter

import (
	"encoding/json"
	api "k8s.io/client-go/pkg/api/v1"
	client "k8s.io/client-go/kubernetes"
)

type KubernetesServiceRepository struct {
	Client client.Interface
}

func (repository *KubernetesServiceRepository) Update(service *api.Service) (*api.Service, error) {
	return repository.Client.CoreV1().Services(service.ObjectMeta.Namespace).UpdateStatus(service)
}

type KubernetesReverseProxyHostConfiguration struct {
	Host string `json:"host"`
}

type KubernetesReverseProxyConfiguration struct {
	Hosts []KubernetesReverseProxyHostConfiguration `json:"hosts"`
}

func GetDomainNamesFromServiceAnnotations(service *api.Service) []string {
	domainNames := []string{}

	value, found := service.ObjectMeta.Annotations["kubernetesReverseproxy"]
	if !found {
		return domainNames
	}

	configuration := KubernetesReverseProxyConfiguration{}
	err := json.Unmarshal([]byte(value), &configuration)
	if err != nil {
		return domainNames
	}

	for _, host := range configuration.Hosts {
		domainNames = append(domainNames, host.Host)
	}

	return domainNames
}
