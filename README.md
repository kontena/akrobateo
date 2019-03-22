# Akrobateo

Akrobateo is a simple [Kubernetes](https://kubernetes.io/) [operator](https://github.com/operator-framework/operator-sdk) to expose in-cluster `LoadBalancer` services as node ports using `DaemonSet`s. The operator naturally also syncs the addresses for the services. This essentially makes the `LoadBalancer` type services behave pretty much like `NodePort` services. The drawback with `NodePort` services is that we're not able to use additional components such as [ExternalDNS](https://github.com/kubernetes-incubator/external-dns) and others.

The node-port proxy Pods utilize iptables to do the actual traffic forwarding.

## Inspiration

This operator draws heavy inspiration from [K3S](https://github.com/rancher/k3s) `servicelb` controller: https://github.com/rancher/k3s/blob/master/pkg/servicelb/controller.go

As K3S controller is fully and tightly integrated into K3S, with good reasons, we thought we'd separate the concept into generic operator usable in any Kubernetes cluster.

## Why `DaemonSet`s?

Running the "proxies" as `DaemonSet`s makes the proxy not to be a single-point-of-failure. So once you've exposed the service you can safaly e.g. push the services external addresses into your DNS. This does have the drawback that a given port can be exposed only in one service throughout the cluster.

## Building

Use the included `build.sh` script. There's naturally also a `Dockerfile` for putting everything into an image.

## Running locally

Either use operator-sdk to run it like so:
```sh
operator-sdk up local
```

Or use the locally built binary:
```sh
WATCH_NAMESPACE="default" ./output/akrobateo_darwin_amd64
```

## Deploying

To deploy to live cluster, use manifests in `deploy` directory. It sets up the operator in `kube-system` namespace with proper service-account and RBAC to allow only needed resources.


## Future

Some ideas how to make things more configurable and/or future-proof

### DaemonSet vs. Deployment

The original Klippy controller creates Deployments. Maybe user could put some annotation on the service whether he/she wants a deployment or a daemonset created. Not really sure how operator-sdk will handle the downstream objects if we create multiple kinds...

### Node selection

There should be some way for the user to select which nodes should act as LBs. So something like a node selector is needed on the services as annotation. That probably also means we'd need to support also tolerations.
