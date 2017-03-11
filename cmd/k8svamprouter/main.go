package main

import (
	"log"
	"os"

	api "k8s.io/client-go/pkg/api/v1"
	client "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	k8svamprouter "github.com/sroze/kubernetes-vamp-router"
)

func main() {
	client := CreateClusterClient()
	serviceUpdater := CreateServiceUpdater(client)

	w, err := client.CoreV1().Services(api.NamespaceAll).Watch(api.ListOptions{
		LabelSelector: labels.Everything(),
		FieldSelector: fields.Everything(),
	})

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
			if (k8svamprouter.ShouldUpdateServiceRoute(service)) {
				serviceUpdater.UpdateServiceRouting(service)
			}
		} else if event.Type == watch.Deleted {
			serviceUpdater.RemoveServiceRouting(service)
		}
	}
}

func CreateClusterClient() client.Interface {
	clusterAddress := os.Getenv("CLUSTER_API_ADDRESS")
	if clusterAddress == "" {
		log.Fatalln("You need to precise the address of Kubernetes API with the `CLUSTER_API_ADDRESS` environment variable")
	}

	config := rest.Config{
		Host: clusterAddress,
		Insecure: os.Getenv("INSECURE_CLUSTER") == "true",
	}

	c, err := client.NewForConfig(&config)
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

func CreateServiceUpdater(client client.Interface) *k8svamprouter.ServiceUpdater {
	rootDns := os.Getenv("ROOT_DNS_DOMAIN")
	if rootDns == "" {
		log.Fatalln("You need to precise your root DNS name with the `ROOT_DNS_DOMAIN` environment variable")
	}

	return &k8svamprouter.ServiceUpdater{
		ServiceRepository: &k8svamprouter.KubernetesServiceRepository{
			Client: client,
		},
		RouterClient: CreateRouterClient(),
		Configuration: k8svamprouter.Configuration{
			RootDns: rootDns,
		},
	}
}
