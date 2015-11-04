package main

import (
	"strings"

	"k8s.io/kubernetes/pkg/api"
)

func (su *ServiceUpdater) GetServiceRouteName(service *api.Service) string {
	return strings.Join([]string{
		service.ObjectMeta.Name,
		service.ObjectMeta.Namespace,
	}, ".")
}

func (su *ServiceUpdater) GetDomainNameFromService(service *api.Service) string {
	return strings.Join([]string{
		su.GetServiceRouteName(service),
		su.Configuration.RootDns,
	}, ".")
}
