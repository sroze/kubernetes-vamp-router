package k8svamprouter

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"errors"
	"fmt"
	"github.com/sroze/kubernetes-vamp-router/vamprouter"

	api "k8s.io/client-go/pkg/api/v1"
)

func GetServiceRouteName(service *api.Service) string {
	return GetDNSIdentifier(strings.Join([]string{
		service.ObjectMeta.Name,
		service.ObjectMeta.Namespace,
	}, "."))
}

func GetDNSIdentifier(name string) string {
	if len(name) > 63 {
		nameHash := GetMD5Hash(name)[0:10]
		name = name[0:52] + "-" + nameHash
	}

	return name
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
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
