package main

import (
	"log"
	"os"

	"k8s.io/kubernetes/pkg/api"
	client "github.com/kubernetes/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/sroze/kubernetes-vamp-router/vamprouter"

)

func main() {
	client := CreateClusterClient()
	serviceUpdater := CreateServiceUpdater(client)

	w, err := client.Services(api.NamespaceAll).Watch(labels.Everything(), fields.Everything(), api.ListOptions{})
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
				serviceUpdater.UpdateServiceRouting(service)
			}
		} else if event.Type == watch.Deleted {
			serviceUpdater.RemoveServiceRouting(service)
		}
	}
}

func CreateClusterClient() *client.Client {
	clusterAddress := os.Getenv("CLUSTER_API_ADDRESS")
	if clusterAddress == "" {
		log.Fatalln("You need to precise the address of Kubernetes API with the `CLUSTER_API_ADDRESS` environment variable")
	}

	config := client.Config{
		Host: clusterAddress,
		Insecure: os.Getenv("INSECURE_CLUSTER") == "true",
	}

	c, err := client.New(&config)
	if err != nil {
		log.Fatalln("Can't connect to Kubernetes API:", err)
	}

	return c
}

func CreateRouterClient() *vamprouter.Client {
	routerAddress := os.Getenv("ROUTER_API_ADDRESS")
	if routerAddress == "" {
		log.Fatalln("You need to precise the address of Vamp Router API with the `ROUTER_API_ADDRESS` environment variable")
	}

	return &vamprouter.Client{
		URL: routerAddress,
	}
}

func CreateServiceUpdater(client *client.Client) *ServiceUpdater {
	rootDns := os.Getenv("ROOT_DNS_DOMAIN")
	if rootDns == "" {
		log.Fatalln("You need to precise your root DNS name with the `ROOT_DNS_DOMAIN` environment variable")
	}

	return &ServiceUpdater{
		ClusterClient: client,
		RouterClient: CreateRouterClient(),
		Configuration: Configuration{
			RootDns: rootDns,
		},
	}
}
