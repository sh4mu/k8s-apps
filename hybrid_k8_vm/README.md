# Exposing k8 services

The goal here is to explore k8s services using minikube, and inter-connect services running on a k8s cluster with services running as on a dedicated VM.

## Create test k8 cluster

Create a minikube k8s cluster on a dedicated VM.

```
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
minikube start
echo 'alias kubectl="minikube kubectl --"' >> ~/.bash_profile
```

## Expose service locally

Create a sample nginx-deployment.yaml and expose it on a NodePort service. 
```
kubectl apply -f nginx-deployment.yaml
kubectl expose deployment nginx-deployment --port=80 --type=NodePort
```

This is only visible within the k8s cluster interface, and the url can be obtained via `minikube service nginx-deployment --url`

## Expose service to remote VM server

A NodePort service is accessible on the minikube k8s network, which is accessible from the host. However this service is not reachable from the outside.

### Use kubectl port forward

This is a simple solution that enables exposing a service, e.g. a service frontend, to the outside world.

On a separate window run the kubectl port-forward `kubectl port-forward --address 10.170.4.1 service/nginx-deployment 8080:80`

Allow 8080 port `sudo firewall-cmd --add-port=8080/tcp`

From an external server, 'curl 10.170.4.1:8080'

The limitation here is that we are simply exposing a service to an external IP, and not creating an actual k8s service exposed to other services. To do so, we need skupper

### Expose services between minikube and external VM using skupper

#### **Expose k8s service to remote server**

We could use different namespace (relevant if we wanted multiple k8s clusters), but in this case we can use the default.

Install and configure skupper
```
curl https://skupper.io/install.sh | sh
export KUBECONFIG=$HOME/.kube/config-west
minikube update-context
```

Start tunnel on the backgroud `minikube tunnel &`

Initiate skupper on this namespace, and create a gateway over podman
```
skupper init
skupper gateway init --type podman
```

Deploy the frontend application `kubectl create deployment frontend --image quay.io/skupper/hello-world-frontend`

Create a skupper service binded to the deployment/frontend, and forward it to the localhost port 8080 using the gateway
```
skupper service create frontend 8080
skupper service bind frontend deployment/frontend
skupper gateway forward frontend 8080
```

Allow 8080 port `sudo firewall-cmd --add-port=8080/tcp --permanent`

Validate that `curl http://localhost:8080` is reachable from the host, and outside the host using the external interface.

#### **Bind remote server service to k8s cluster**

The goal is to run the backend service on a VM and make it visible via skupper.

Clone the backend application code, `git clone https://github.com/skupperproject/skupper-example-gateway.git`

Install pip, to install python3 missing modules. Need `python39-pip-20.2.4-7.module+el8.6.0+13003+6bb2c488.noarch.rpm`.

Run backend on external interface `$ python3 python/main.py --host 10.110.1.30 --port 8080) &`

On the minikube host validate connectivity, e.g.  `$ curl http://10.110.1.30:8080`

Create skupper backend service, and bind it to service remoter server IP and Port

```
skupper service create backend 8080
skupper gateway bind backend 10.110.1.30 8080
```

Run `curl http://10.110.4.1:8080/api/health` from anywhere and expect `OK`


# Implement Prometheus example

The goal is to extend the first attempt, and deploy a prometheus environment capable of collecting metrics from services running in different namespaces and on dedicated VMs outside the k8 host.

## Expose prometheus application on skupper gateway

Install prometheus using the official helm charts. It will install on the "monitoring" namespace
```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/prometheus
```

Change to the monitoring namespace `kubectl config set-context --current --namespace monitoring`

If we were to expose it to other application running on a different namespace, the "standard" procedure would be. **DON'T RUN THIS**
`$ kubectl expose service prometheus-server --type=NodePort --target-port=9090 --name=prometheus-server-np`

However we do not want a NodePort service. We want to create a service via skupper and forward it to the localhost using the skupper gateway.

Start by removing the ClusterIP service associated to prometheus-server
`$ kubectl delete service/prometheus-server`

Create a skupper service on port 80
`$ skupper service create prometheus-server 80`

Bind this service to the prometheus-server application running on port 9090
`$ skupper service bind prometheus-server deployment/prometheus-server --target-port 9090`

Forward the prometheus-server service to the localhost port 8080
`$ skupper gateway forward prometheus-server 8080`

voila! `curl http://127.0.0.1:8080`

## Connect prometheus with metrics running on different namespaces

Clone the skupper example for prometheus `git clone https://github.com/skupperproject/skupper-example-prometheus.git`

Start minikube and `minikube tunnel`

This example connects different clusters, represented by different namespaces. Just implement the public one from the example.

Start skupper on public1 namespace and create token for cluster connectivity; on *terminal1*
```
kubectl create namespace public1
kubectl config set-context --current --namespace public1
skupper init
skupper token create public1-token.yaml --uses 2
```

Start skupper on public2 namespace and connect to public1
```
kubectl create namespace public2
kubectl config set-context --current --namespace public2
skupper init
skupper token create public2-token.yaml
skupper link create public1-token.yaml

skupper link status
```

Add deployment for metric generation on *terminal1* `kubectl apply -f metrics-deployment-b.yaml`

Add deployment for prometheus on *terminal2* `kubectl apply -f prometheus-deployment.yaml`

On *terminal1* Expose metric service and add `app=metrics` label that is recognized by prometheus configuration
```
skupper expose deployment metrics-b --address metrics-b --port 8080 --protocol tcp --target-port 8080
skupper service label metrics-b app=metrics
```

On *terminal2* expose prometheus service
```
skupper expose deployment prometheus --address prometheus --port 9090 --protocol http --target-port 9090
```

Confirm that it is available, `curl http://$(kubectl get service prometheus -o=jsonpath='{.spec.clusterIP}'):9090`

To expose the prometheus to the host external interface, forward the prometheus service, note
* There was an issue with the gateway, so I had to rebuild it
* forward to 9090, do not use 8080 (it goes to some skupper hello world..)

```
skupper gateway delete
skupper gateway init --type podman
skupper gateway forward prometheus 8080
```

## Connect prometheus with metrics running on different server

Add a new service to the public1 namespace to bind to the external server service
```
skupper service create metrics-timesten 8081
skupper gateway bind metrics-timesten 10.140.1.30 9090
```

Add the label, so that the prometheus automatically updates its target `skupper service label metrics-timesten app=metrics`



