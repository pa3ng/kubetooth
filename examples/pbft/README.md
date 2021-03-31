# sawtooth-pbft-example

## Prerequisites

1. Docker Desktop installed
2. Kubernetes enabled on Docker Desktop
3. Ensure you're on the `docker-destkop` context
   ```bash
   $ kubectl config use-context docker-desktop
   ```
4. Clone this repo

## Deployment

1. Run the `sawtooth-create-pbft-keys.yaml` job to generate key pairs for the 5 sawtooth nodes
2. Copy and paste the key pair values in the `pbft-keys-configmap.yaml`
3. Create the configmap resource by
  ```bash
  $ kubectl apply -f pbft-keys-configmap.yaml
  ```
4. Deploy the sawtooth services
 ```bash
  $ kubectl apply -f service0.yaml -f service1.yaml -f service2.yaml -f service3.yaml -f service4.yaml
  ```
5. Deploy the genesis node first, and check the logs to see if has successfully initialized the genesis block
  ```bash
  $ kubectl apply -f node0.yaml
  $ kubectl logs pbft-0-hash_of_pod
  ```
6. Deploy the peer nodes and check for successful connections in the genesis node logs
  ```bash
  $ kubectl apply -f node1.yaml -f node2.yaml -f node3.yaml -f node4.yaml
  ```
7. Check for successful peer connection logs in the genesis node. For example:
  ```bash
  $ kubectl logs pbft-0-hash_of_pod
  [2020-11-06 17:57:24.612 DEBUG  gossip] Added connection_id 0f8591f6a83ecbb297dbd0dfb27d10b8dcd8dfd51728d1a47a0d084304c142d4bcc0280c7c457968ac17dd0f8a7468fbbb34c8e891abcd33678ca7bdb10c1f42 with endpoint tcp://10.105.110.136:8800, connected identities are now {'0f8591f6a83ecbb297dbd0dfb27d10b8dcd8dfd51728d1a47a0d084304c142d4bcc0280c7c457968ac17dd0f8a7468fbbb34c8e891abcd33678ca7bdb10c1f42': 'tcp://10.105.110.136:8800'}
  ```

## Setup Kubernetes Dashboard

The Dashboard UI is not deployed by default. To deploy it, run the following command:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0/aio/deploy/recommended.yaml
```

You can now access Dashboard using the kubectl command-line tool by running the following command:

```bash
$ kubectl proxy
```

And then navigating to `http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy`

However, you will need to first authenticate with a valid Bearer Token to log in to the dashboard. To obtain this Bearer Token, we can simply borrow a token from an existing Service Account.

```bash
$ kubectl -n kube-system get secret
NAME                                             TYPE                                  DATA   AGE
default-token-5t7rv                              kubernetes.io/service-account-token   3      24h
deployment-controller-token-z56l5                kubernetes.io/service-account-token   3      24h
```

You'll likely see many Service Accounts, but we only need to access one of them.

```bash
$ kubectl -n kube-system describe secret deployment-controller-token-z56l5
```

We'll borrow the Bearer Token from the deployment controller by copying the value located in the `token` field.

Now, paste that value in the `Enter token *` field in the UI. Click the sign in button and you're good to go.