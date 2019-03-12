# Service LB Operator

Simple operator to spin-up [klippy-lb]() DaemonSets and sync the addresses for the services.

## Building

Use the included `build.sh` script. There's naturally also a `Dockerfile` for putting everything into an image.

## Running locally

Either use operator-sdk to run it like so:
```sh
operator-sdk up local
```

Or use the locally built binary:
```sh
WATCH_NAMESPACE="default" ./output/manager_darwin_amd64
```

## Deploying

To deploy to live cluster, use manifests in `deploy` directory. It sets up the operator in `kube-system` namespace with proper service-account and RBAC to allow only needed resources.

## TODO items

- test with multiple services ports
- test when updating the service
- docs needed

## Future

Some ideas how to make things more configurable and/or future-proof

### DaemonSet vs. Deployment

The original Klippy controller creates Deployments. Maybe user could put some annotation on the service whether he/she wants a deployment or a daemonset created. Not really sure how operator-sdk will handle the downstream objects if we create multiple kinds...

### Node selection

There should be some way for the user to select which nodes should act as LBs. So something like a node selector is needed on the services as annotation. That probably also means we'd need to support also tolerations
