# service-lb-operator

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

## TODO items

- automated builds needed
- test with multiple services ports
- test when updating the service
- deployment yamls
- docs needed

## Future

Some ideas how to make things more configurable:

### DaemonSet vs. Deployment

The original Klippy controller creates Deployments. Maybe user could put some annotatino on the service whether he/she wants a deployment or a daemonset created. Not really sure how operator-sdk will handle the downstream objects if we create multiple kinds...