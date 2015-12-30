package k8svamprouter

import (
	"strings"
	"crypto/md5"
	"encoding/hex"

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
		name = name[0:52]+"-"+nameHash
	}

	return name
}

func (su *ServiceUpdater) GetDomainNameFromService(service *api.Service) string {
	return strings.Join([]string{
		GetServiceRouteName(service),
		su.Configuration.RootDns,
	}, ".")
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
