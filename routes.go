package k8svamprouter

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"k8s.io/kubernetes/pkg/api"
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

func (su *ServiceUpdater) GetDomainNamesFromService(service *api.Service) []string {
	domainNames := GetDomainNamesFromServiceAnnotations(service)

	// Add the default domain name
	domainNames = append(domainNames, strings.Join([]string{
		GetServiceRouteName(service),
		su.Configuration.RootDns,
	}, "."))

	return domainNames
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
