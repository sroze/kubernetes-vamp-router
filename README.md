# Kubernetes Vamp Router

When dealing with your own Kubernetes Cluster (not hosted in GCE or AWS) you will have trouble to exposes automatically
public services. With these cloud providers, we simply have to create a LoadBalancer service and just wait a bit to have
an address automatically added in the status of the service.

This bridge for Vamp Router will automatically creates a new route via the [Vamp Router](https://github.com/magneticio/vamp-router)
API when a service is created or updated (if needed). As well, when the service is removed, then the route is removed.

![Overview of kubernetes-vamp-router in an architecture](docs/architecture.png)

*Features:*

- Automatically creates routes on Vamp Router when a `LoadBalancer` service is created
- Updates the service's status to declare the created route
- Read the annotations to create custom hosts

## Getting started

First of all, start your Vamp router on your front-end server:
```
docker run -d \
    --name=vamp-router \
    --net=host \
    --restart=always \
    magneticio/vamp-router:latest
```

Then, you just have to start the container `sroze/kubernetes-vamp-router` with the right environment variables based
on your configuration. [Be careful about where you run these containers, they need to have access to the running containers](#where-to-run-these-containers).

```
docker run -d \
    --restart=always \
    --name=kubernetes-vamp-router \
    -e ROOT_DNS_DOMAIN=any.wildcarded.dns.address \
    -e CLUSTER_API_ADDRESS=https://username:password@your.master.cluster.address \
    -e INSECURE_CLUSTER=true \
    -e ROUTER_API_ADDRESS=http://your.router.api.address:10001/
    sroze/kubernetes-vamp-router
```

## Configuration

The configuration can be done by passing some environment variables. Here is the list and the meaning of all environment
variables that the container is reading:

- `CLUSTER_API_ADDRESS`: Address of the Kubernetes cluster API
- `ROUTER_API_ADDRESS`: Address of the Vamp Router API
- `ROOT_DNS_DOMAIN`: The root DNS address that needs to be used. Basically, this (sub)domain name should be a wildcard
  to your front-end server on which your Vamp Server is running.
- `INSECURE_CLUSTER`: If the value is `true`, then the SSL certification won't be checked. This should be used for
  development purposes only!

## Where to run these containers?

You have to run them on a machine which that:
- Have kube-proxy running (to be able to connect to pods using services' IPs)
- Is in the cluster network (to be able to actually route traffic to running containers)

The easiest way is to run them on a public node of your cluster but running them outside just requires you to configure the networking and install kube-proxy.



