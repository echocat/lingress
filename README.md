# lingress

lingress = The lean ingress controller.

In contrast to the most other ingress controllers for Kubernetes, this one does not come with extreme many different features and tries to be extreme dynamic. It tries to be as [lean](https://en.wikipedia.org/wiki/Lean_software_development) as possible.

## Topics

1. [Main principles](#main-principles)
2. [Requirements](#requirements)
3. [Setup](doc/setup.md) ➡️
4. [Configuration](doc/configuration.md) ➡️
5. [Headers](doc/headers.md) ➡️
6. [Examples](doc/examples.md) ➡️
7. [Contributing](CONTRIBUTING.md) ➡️
8. [Code of Conduct](CODE_OF_CONDUCT.md) ➡️
9. [License](LICENSE) ➡️

## Main principles

1. **Minimal configuration**: Just one single configuration and this as short as possible; without repeating common stuff for all ingress configurations (like ensure CORS). The majority of the ingress controllers coming nowadays with other many several configurations and/or you have to define different settings for http and https.<br><br>

2. **Centralize standards**: Settings like [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) or [HSTS](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) usually settings you want to have for your whole application. For the majority of the ingress controllers, these settings are made per ingress configuration. lingress provides the feature to centralize these settings and even [force them](#forcible).<br><br>

3. **HTTPS by default**: As lingress makes no difference between http and https by default, it also forces (which is nowadays industry standards) the clients to https (if the clients comes from http).<br><br>

4. **Industry standards by default**: Instead of always activating some industry standards or even need plugins to add them lingress, supports the important ones out of the box. See [headers sections](#headers) for examples. 

## Requirements

1. [Kubernetes cluster with API v1.22.0+](https://kubernetes.io/releases/) (or API compatible such as [k3s](https://k3s.io/) or [OpenShift Kubernetes Engine](https://docs.openshift.com/container-platform/latest/welcome/oke_about.html))
2. [Helm v3.0.0+](https://helm.sh/) 

## More topics
* [Setup](doc/setup.md)
* [Configuration](doc/configuration.md)
* [Examples](doc/examples.md)
* [Headers](doc/headers.md)
