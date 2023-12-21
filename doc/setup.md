# lingress Setup

[lingress](../README.md) can currently be setup out-of-the-box [with Helm](#with-helm).

## With [Helm](https://helm.sh/)

1. #### Register the [echocat Helm chart](https://packages.echocat.org/helm):
   ```shell
   helm repo add echocat https://packages.echocat.org/helm
   ```

2. #### Install/Upgrade the chart:

   ##### Basics
   ```shell
   helm upgrade --install --atomic -n kube-system lingress echocat/lingress
   ``` 

   ##### With parameters
   Providing parameters using `--set <param>=<value>`, `--set-string <param>=<value>` or `--set-json <param>=<value>` (see [Helm Upgrade documentation](https://helm.sh/docs/helm/helm_upgrade/) for more details). The following example ensures, that the helm is never deleting the service of lingress:
   ```shell
   helm upgrade --install --atomic -n kube-system lingress echocat/lingress \
        --set-json 'service.annotations={"helm.sh/resource-policy":"keep"}'
   ```

   See all supported values: [charts/lingress/values.yaml](../charts/lingress/values.yaml) and [configuration in general](configuration.md).

## More topics
* [Configuration](configuration.md)
* [Headers](headers.md)
* [Examples](examples.md)
* [Overview](../README.md)
