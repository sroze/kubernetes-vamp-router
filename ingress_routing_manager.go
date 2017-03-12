package k8svamprouter

import (
	"log"
	"fmt"

	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	client "k8s.io/client-go/kubernetes"
)

type IngressRoutingManagerConfiguration struct {
	RootDns string
	IngressType string
}

type IngressRoutingManager struct {
	Configuration IngressRoutingManagerConfiguration

	KubernetesClient client.Interface
}

func (irm *IngressRoutingManager) ShouldHandleObject(object KubernetesBackendObject) bool {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		log.Println("[error] Get get only from `Ingress` objects")

		return false
	}

	if irm.Configuration.IngressType != ingress.Annotations["kubernetes.io/ingress.class"] {
		return false
	}

	return true
}

func (irm *IngressRoutingManager) GetDomainNames(object KubernetesBackendObject) ([]string, error) {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		return nil, fmt.Errorf("Get get only from `Ingresss` objects")
	}

	domainSeparator := GetDomainSeparator()

	domainNames := []string{
		GetRouteNameFromObjectMetadata(ingress.ObjectMeta, domainSeparator)+irm.Configuration.RootDns,
	}

	return domainNames, nil
}

func (irm *IngressRoutingManager) GetRouteName(object KubernetesBackendObject) (string, error) {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		return "", fmt.Errorf("Get get only from `Ingress` objects")
	}

	return GetRouteNameFromObjectMetadata(ingress.ObjectMeta, GetDomainSeparator()), nil

}

func (irm *IngressRoutingManager) GetBackendAddress(object KubernetesBackendObject) (string, error) {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		return "", fmt.Errorf("Get get only from `Ingress` objects")
	}

	return ingress.Spec.Backend.ServiceName+"."+ingress.ObjectMeta.Namespace+".svc.cluster.local", nil
}

func (irm *IngressRoutingManager) UpdateObjectWithDomainNames(object KubernetesBackendObject, domainNames []string) error {
	ingress, ok := object.(*v1beta1.Ingress)
	if !ok {
		return fmt.Errorf("Get get only from `Ingress` objects")
	}

	ingress.Status = v1beta1.IngressStatus{
		LoadBalancer: CreateLoadBalancerStatusFromDomainNames(domainNames),
	}

	_, err := irm.KubernetesClient.ExtensionsV1beta1().Ingresses(ingress.ObjectMeta.Namespace).UpdateStatus(ingress)

	return err
}
