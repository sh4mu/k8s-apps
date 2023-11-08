# Create test k8 cluster

Create a VM on vCenter and add an interface on VLAN170 (testing).

On the OS run,
```
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
minikube start
alias kubectl="minikube kubectl --"
```

# Exposing k8 services

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
