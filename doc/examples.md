# lingress Examples

On this page you can find a selection of some examples to get easier onboard with [lingress](../README.md).

> [!NOTE]
> Please refer [setup](setup.md) and [configuration](configuration.md) for a better context.

## Load balancer with PROXY protocol

> [!NOTE]
> Usually an external load balancer is used. In those cases lingress will not be connected directly to the clients. Therefore, lingress will see as the remote IP, the IP of the load balancer and not of the client itself. To solve that, many load balancers implements the [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt), which is also supported by lingress.

To enable the [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) support, we need:
1. Tell the load balancer to wrap the TCP packages accordingly. Be aware: Each load balancer works differently. Study the according documentation.
2. Tell lingress to evaluate them.

* With [AWS EKS](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/guide/service/annotations/#proxy-protocol-v2):
  ```shell
  helm upgrade --install --atomic -n kube-system lingress echocat/lingress \
      --set-json 'service.annotations={"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes":"proxy_protocol_v2.enabled=true"}' \
      --set-json 'controller.args=["--server.http.proxyProtocol.respect","--server.https.proxyProtocol.respect"]'
  ```
* With [OVHcloud Managed Kubernetes](https://help.ovhcloud.com/csm?id=kb_article_view&sysparm_article=KB0050019):
  ```shell
  helm upgrade --install --atomic -n kube-system lingress echocat/lingress \
      --set-json 'service.annotations={"service.beta.kubernetes.o/ovh-loadbalancer-proxy-protocol":"v2"}' \
      --set-json 'controller.args=["--server.http.proxyProtocol.respect","--server.https.proxyProtocol.respect"]'
  ```

## Using TLS certificate secrets

> [!NOTE]
> By default lingress will **NOT** have any certificates for TLS configured. Assuming now you have [cert-manager](https://cert-manager.io/) part of your cluster to managing certificates for you.

1. Have [cert-manager](https://cert-manager.io/docs/installation/helm/) installed.
2. Have a certificate inside your cluster, like:
   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: my-tls-ceritificate
     namespace: my-namespace
   spec:
     dnsNames:
       - my-domain.org
     issuerRef:
       kind: ClusterIssuer
       name: my-issuer
     secretName: my-tls-ceritificate
     secretTemplate:
       labels:
         # This label will we be afterward part the created secret.
         # ...and will tell lingress to find these secrets. 
         my-public-tls-certificates: "true"
   ```
3. Configure lingress accordingly
   ```shell
   helm upgrade --install --atomic -n kube-system lingress echocat/lingress \
      --set-json 'controller.args=["--tls.secretLabelSelector=my-public-tls-certificates=true"]'
   ```

## Using DaemonSets

> [!NOTE]
> By default lingress runs with [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/). This is great by default. You define the Helm value `controller.replicas=<amount>` and lingress will run at these amount of replicas. Also rolling updates are working out-of-the-box. But in some scenarios you want that each [node](https://kubernetes.io/docs/concepts/architecture/nodes/) runs its own instance. In can improve the latency of all requests.

  ```shell
  helm upgrade --install --atomic -n kube-system lingress echocat/lingress \
      --set-string 'controller.kind=DaemonSet'
  ```

## More topics
* [Setup](setup.md)
* [Configuration](configuration.md)
* [Headers](headers.md)
* [Overview](../README.md)
