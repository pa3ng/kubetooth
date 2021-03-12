# kubetooth

Status: **WIP**

Easily deploy a hyperledger sawtooth blockchain to a kubernetes cluster

### Quick Start

Ensure docker-desktop is running with kubernetes enabled

Build the image

```sh
make docker-build BIN=kubetooth
```

Run the container

```sh
docker run -v "$PWD"/config:/config -v ~/.kube:/.kube kubetooth:local
```

Verify deployment by listing your pods

```
kubectl get pods
```

Should return something like this

```
NAME                                          READY   STATUS    RESTARTS   AGE
conensource-sawtooth-pbft-0-895d48975-l69qx   5/5     Running   0          1m12s
```