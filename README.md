# Sentry + Opentelemetry Go Example

## Requirements
To run this example, you will need a kubernetes cluster.
This example has been tried and tested on
- Minikube
- AWS Elastic Kuberetes Service
and should work on any configured kubernetes environment.

## K8S components used
- Config Map (otel-collector-conf) for opentelemetry collector
- Service to expose port 4317 of collector deployment (to recieve OTLP traces)
- Deployment
  - Opentelemetry Collector
  - Instrumented Golang application Deployment
 
## Steps 
1. Create an `observabilitiy` namespace in the cluster.

```cmd
$ kubectl create ns observability
```

2. Change the configmap in `k8s.yaml` and add your sentry DSN in line 31

```yaml
      sentry:
        dsn: <your sentry DSN here>
```
3. Apply the kubernetes configuration in the observability namespace.
```cmd
$ kubectl apply -f k8s.yaml -n observability
```

4. Once the deployment is done, port forward the sample gin server to your localhost (ensure that your 8088 port is free from any pre-binding, i.e. no other server is using that)

```cmd
$ kubectl port-forward -n observability svc/otlp-instrumentation-demo 8088:8088
```

5. Visit your [localhost:8088/users/123](http://localhost:8088/users/123) 

6. Visit your Sentry UI to see something similar to this
![Sentry UI](./img/sentryui.png)
