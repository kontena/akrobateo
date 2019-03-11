
This folder contains a simple service to test the operator with.

## Results

After deploying the pods and service, you should see something like this:
```
$ kubectl get svc
NAME               TYPE           CLUSTER-IP       EXTERNAL-IP                                                             PORT(S)                         AGE
echoserver         LoadBalancer   172.32.196.236   147.75.100.201,147.75.101.127,147.75.102.57,147.75.85.43,147.75.85.45   8080:31278/TCP,9090:30987/TCP   34m
kubernetes         ClusterIP      172.32.0.1       <none>                                                                  443/TCP                         5d3h
```

You should be able to access the service on given ports through each of the nodes:
```
$ curl 147.75.85.43:9090
CLIENT VALUES:
client_address=('172.31.5.104', 58861) (172.31.5.104)
command=GET
path=/
real path=/
query=
request_version=HTTP/1.1

SERVER VALUES:
server_version=BaseHTTP/0.6
sys_version=Python/3.5.0
protocol_version=HTTP/1.0

HEADERS RECEIVED:
Accept=*/*
Host=147.75.85.43:9090
User-Agent=curl/7.57.0
```

Given of course that there's no firewalls or such blocking those ports.