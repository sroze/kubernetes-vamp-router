package main

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	client "github.com/kubernetes/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/watch"

	"os"
)

func main() {
	clusterAddress := os.Getenv("CLUSTER_ADDRESS")
	rootDns := os.Getenv("ROOT_DNS_DOMAIN")
	if rootDns == "" {
		log.Fatalln("You need to precise your root DNS name with the `ROOT_DNS_DOMAIN` environment variable")
	}

	config := client.Config{
		Host: clusterAddress,
		Insecure: os.Getenv("INSECURE_CLUSTER") == "true",
	}

	c, err := client.New(&config)
	if err != nil {
		log.Fatalln("Can't connect to Kubernetes API:", err)
	}

	w, err := c.Services(api.NamespaceAll).Watch(labels.Everything(), fields.Everything(), api.ListOptions{})
	if err != nil {
		log.Fatalln("Unable to watch services:", err)
	}

	log.Println("Watching services")
	for event := range w.ResultChan() {
		service, ok := event.Object.(*api.Service)
		if !ok {
			log.Println("Got a non-service object")

			continue
		}

		if event.Type == watch.Added || event.Type == watch.Modified {
			if (ShouldUpdateServiceRoute(service)) {
				UpdateServiceRouting(service)
			}
		} else if event.Type == watch.Deleted {
			RemoveServiceRouting(service)
		}
	}
}

func ShouldUpdateServiceRoute(service *api.Service) bool {
	if service.Spec.Type != api.ServiceTypeLoadBalancer {
		log.Println("Skipping service", service.ObjectMeta.Name, "as it is not a LoadBalancer")

		return false
	}

	// If there's an IP and/or DNS address in the load balancer status, skip it
	if ServiceHasLoadBalancerAddress(service) {
		log.Println("Skipping service", service.ObjectMeta.Name, "as it already have an address")

		return false
	}

	return true
}

func ServiceHasLoadBalancerAddress(service *api.Service) bool {
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return false
	}

	for _, ingress := range service.Status.LoadBalancer.Ingress {
		if ingress.IP != "" || ingress.Hostname != "" {
			return true
		}
	}

	return false
}

func UpdateServiceRouting(service *api.Service) {
	log.Println("Should update service routing")
}

func RemoveServiceRouting(service *api.Service) {
	log.Println("Should remove service routing")
}

