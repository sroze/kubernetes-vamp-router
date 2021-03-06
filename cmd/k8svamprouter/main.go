package main

import (
	"log"
	"os"
	"sync"

	api "k8s.io/client-go/pkg/api/v1"
	client "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"

	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/watch"

	"github.com/sroze/kubernetes-vamp-router/vamprouter"
	k8svamprouter "github.com/sroze/kubernetes-vamp-router"
)

func main() {
	client := CreateClusterClient()

	messages := make(chan int)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if "yes" == os.Getenv("WATCH_SERVICES") {
			WatchServices(client)
		}

		messages <- 1
	}()

	go func() {
		defer wg.Done()

		watchIngresses := os.Getenv("WATCH_INGRESSES")
		if "" == watchIngresses || "yes" == watchIngresses {
			go WatchIngresses(client)
		}

		messages <- 1
	}()

	wg.Wait()
}

func WatchIngresses(kubernetesClient client.Interface) {
	log.Println("Watching Kubernetes ingresses")

	ingressType := os.Getenv("INGRESS_TYPE")
	if "" == ingressType {
		ingressType = "vamp-router"
	}

	routingManager := CreateRouteManager(
		&k8svamprouter.IngressRoutingManager{
			KubernetesClient: kubernetesClient,
			Configuration: k8svamprouter.IngressRoutingManagerConfiguration{
				RootDns: os.Getenv("ROOT_DNS_DOMAIN"),
				IngressType: ingressType,
			},
		},
	)

	channel, err := kubernetesClient.ExtensionsV1beta1().Ingresses(api.NamespaceAll).Watch(api.ListOptions{
		LabelSelector: labels.Everything().String(),
		FieldSelector: fields.Everything().String(),
	})

	if err != nil {
		log.Fatalln("Unable to watch ingresses:", err)
	}

	WatchObjects(routingManager, channel)
}

func WatchServices (kubernetesClient client.Interface) {
	log.Println("Watching Kubernetes services")

	routeManager := CreateRouteManager(
		CreateServiceUpdater(kubernetesClient),
	)

	channel, err := kubernetesClient.CoreV1().Services(api.NamespaceAll).Watch(api.ListOptions{
		LabelSelector: labels.Everything().String(),
		FieldSelector: fields.Everything().String(),
	})

	if err != nil {
		log.Fatalln("Unable to watch services:", err)
	}

	WatchObjects(routeManager, channel)
}

func WatchObjects(routeManager *k8svamprouter.VampRouteManager, channel watch.Interface) {
	for event := range channel.ResultChan() {
		if !routeManager.ShouldHandleObject(event.Object) {
			continue
		}

		if event.Type == watch.Added || event.Type == watch.Modified {
			routeManager.UpdateObjectRouting(event.Object)
		} else if event.Type == watch.Deleted {
			routeManager.RemoveObjectRouting(event.Object)
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
		Configuration: k8svamprouter.Configuration{
			RootDns: rootDns,
		},
	}
}

func CreateRouteManager(objectRoutingResolver k8svamprouter.ObjectRoutingResolver) *k8svamprouter.VampRouteManager {
	return &k8svamprouter.VampRouteManager{
		RouterClient: CreateRouterClient(),
		ObjectRoutingResolver: objectRoutingResolver,
	}
}
