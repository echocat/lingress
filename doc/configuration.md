# lingress Configuration

[lingress](../README.md) will be mainly configured by:
1. Command line arguments, which apply to lingress globally (see [Parameters](#parameters)),
2. ... Ingress configuration annotations (see [Parameters](#parameters))
3. ... and the [Helm values](#helm-values).

## TOC

1. [Parameters](#parameters)
   1. [Forcible](#forcible)
1. [Helm values](#helm-values)
    
## Parameters

| CLI Flag | Annotation | Default | [Forcible](#forcible) | Description |
|--|--|---|--|--|
| `--accessLog.queueSize` | | `5000` | | Maximum number of accessLog elements that could be queue before blocking. |
| `--accessLog.inline` | | `true` | | MInstead of exploding the accessLog entries into sub-entries everything is inlined into the root object. |
| `--client.http[s].maxRequestHeaderBytes` | | `2MB` | | Maximum number of bytes the client will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body. |
| `--client.http[s].readHeaderTimeout` | | `30s` | | Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body. |
| `--client.http[s].writeTimeout` | | `30s` | | Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read. |
| `--client.http[s].idleTimeout` | | `5m` | | Maximum amount of time to wait for the next request when keep-alives are enabled. |
| `--client.http[s].keepAlive` | | `2m` | | Duration to keep a connection alive (if required); 0 means unlimited. |
| `--cors.enabled` | `lingress.echocat.org/cors.enabled` | `false` | `L`/`C` | If `true` it will enable [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for the service. |
| `--cors.allowedOriginHosts` | `lingress.echocat.org/cors.allowed-origin-hosts` | | `L`/`C` | Glob pattern to allow origin hosts for [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). `*` means all. |
| `--cors.allowedMethods` | `lingress.echocat.org/cors.allowed-methods` | | `L`/`C` | List of methods to allow for [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). `*` means all. |
| `--cors.allowedHeaders` | `lingress.echocat.org/cors.allowed-headers` | | `L`/`C` | List of headers to allow for [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). `*` means all. |
| `--cors.allowedCredentials` | `lingress.echocat.org/cors.credentials` | `true` | `L`/`C` | `true` means that credentials are allowed for [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). |
| `--cors.maxAge` | `lingress.echocat.org/cors.max-age` | `24h` | `L`/`C` | How long the response to the preflight request can be cached for without sending another preflight request based on [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). |
| `--discovery.resyncAfter` | | `10m` | | How often lingress should execute a full sync of all settings of the Kubernetes cluster. |
| `--fallback.reloadTimeoutOnTemporaryIssues` | | `15s` | | How often the fallback should reload the page on temporary issues. |
| `--ingress.class` | | `lingress,` | | To which ingress classes lingress should handle. |
| `--kubernetes.config` | | `~/.kube/config` | | Defines the location of the configuration to communicate with Kubernetes. If `incluster` it will use the cluster internal configuration. |
| `--kubernetes.context` | | `<default>` | | Defines the context of the configuration to communicate with Kubernetes. In case of `incluster` it will be ignored. |
| `--kubernetes.namespace` | | `<default>` | | Defines the namespace within Kubernetes. In case of `incluster` it will be ignored. |
| `--log.level` | | `info` | | Defines all possible levels to be logged at. Can be `debug`, `info`, `warn`, `error` or `fatal`. |
| `--log.format` | | `json` | | Defines in which format will be logged in. Can be `text` and `json`. |
| `--log.color` | | `auto` | | Defines in which format will be logged in. Can be `auto`, `always` and `never`. |
| `--management.maxRequestHeaderBytes` | | `2MB` | | Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body. |
| `--management.readHeaderTimeout` | | `30s` | | Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body. |
| `--management.writeTimeout` | | `1m` | | Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read. |
| `--management.idleTimeout` | | `5m` | | Maximum amount of time to wait for the next request when keep-alives are enabled. |
| `--management.pprof` | | `false` | | Will serve at the management endpoint pprof profiling, too. DO NOT USE IN PRODUCTION! |
| `--request.headers` | `lingress.echocat.org/headers.request` | | `L`/`C` | Could be defined multiple times (for cli) or separated by `\n` (for annotations) and will set, add(`+`) or remove(`-`) headers going to upstream. Each entry has to be defined by `<name>:<value>`. |
| `--response.headers` | `lingress.echocat.org/headers.response` | | `L`/`C` | Could be defined multiple times (for cli) or separated by `\n` (for annotations) and will set, add(`+`) or remove(`-`) headers going to client. Each entry has to be defined by `<name>:<value>`. |
| `--response.compress` | `lingress.echocat.org/compress.enabled` | `true` | `L` | If `true` each response will be compressed before streaming to the client (if meaningful). |
| `--server.http[s].listenAddress` | | `:8080`/`:8443` | | Address lingress will listen for HTTP(s) requests for. |
| `--server.http[s].maxConnections` | | `256`/`512` | |  Maximum amount of connections handled by lingress concurrently via HTTP(s).|
| `--server.http[s].soLinger` | | `-1` | | Set the behavior of `SO_LINGER`. See [Manpages](https://man7.org/linux/man-pages/man7/socket.7.html), [Stackoverflow](https://stackoverflow.com/questions/3757289/when-is-tcp-option-so-linger-0-required) and [IBM docs](https://www.ibm.com/docs/en/cics-tg-multi/9.2?topic=settings-so-linger-setting) for more information. |
| `--server.http[s].proxyProtocol.respect` | | `false` | | If set to `true` it will respect the [proxy protocol](https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt) to evaluate remote IPs etc. from upstream. Currently version 1&2 is supported. |
| `--server.behindReverseProxy` | | `false` | | If set to `true` it will respect `X-Forwarded` headers to evaluate the remote IPs etc. |
| `--tls.secretNames` | | | | Names of secrets that contains TLS key and certificate pairs. They can be of format `[<namespace>/]<name>`. If no namespace is specified, `--kubernetes.namespace` is used as base. This parameter can be specified multiple times. Together with `--tls.secretNamePatterns` this will act as `OR` combination. |
| `--tls.secretNamePatterns` | | | | Regex pattern to match names of secrets that contains TLS key and certificate pairs. The pattern has to match `<namespace>/<name>`. This parameter can be specified multiple times. Together with `--tls.secretNames` this will act as `OR` combination. |
| `--tls.secretLabelSelector` | | | | [Label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) which all secrets have to met to be eligible as secrets that contains TLS key and certificate pairs. This criteria has to met additionally to all other criteria (`AND` condition). |
| `--tls.secretFieldSelector` | | | | [Field selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/) which all secrets have to met to be eligible as secrets that contains TLS key and certificate pairs. This criteria has to met additionally to all other criteria (`AND` condition). |
| `--tls.forced` | `lingress.echocat.org/force-secure` | `false` | `L` | If `true` each request to `http` will be forcible redirected to `https`. |
| `--tls.fallbackCertificate` | | `false` | | If `true` lingress will respond with a dummy certificate if no matching certificate can be found. Otherwise the TLS handshake will be interrupted. |
| `--upstream.maxIdleConnectionsPerHost` | | `20` | | Controls the maximum idle (keep-alive) connections to keep per-host. |
| `--upstream.maxConnectionsPerHost` | | `250` | | Limits the total number of connections per host, including connections in the dialing, active, and idle states. On limit violation, dials will block. |
| `--upstream.idleConnectionTimeout` | | `1m` | | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit. |
| `--upstream.maxResponseHeaderSize` | | `20MB` | | Limit on how many response bytes are allowed in the server's response header. |
| `--upstream.dialTimeout` | | `10s` | | Maximum amount of time a dial will wait for a connect to complete. If Deadline is also set, it may fail earlier. |
| `--upstream.keepAlive` | | `30s` | | Keep-alive period for an active network connection. If zero, keep-alives are enabled if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alives are disabled. |
| `--upstream.override.host` | | | | Overrides the target host always with this value. Only for testing. |
| `--upstream.override.scheme` | | | | Overrides the target scheme always with this value. Only for testing. |
| | `lingress.echocat.org/strip-rule-path-prefix` | `false` | | If `true` a matched prefix from the ingress rule will be removed. In case of `false` it remain. Example: Rule has `/foo` and request is `/foo/bar`; `false=/foo/bar`; `true=/bar` |
| | `lingress.echocat.org/path-prefix` | | | If provided this path will be always be prepended before sending to the upstream. Example: Request path is `/bar`; `<empty>=/bar`; `/foo=/foo/bar` |
| | `lingress.echocat.org/x-forwarded-prefix` | `true` | | If `true` the upstream will receive an header which contains matched prefix of the ingress rule. |
| | `lingress.echocat.org/whitelisted-remotes` | | | List of IPs and/or host names which are allowed to access the endpoint. `*` can be used. And many entries can be separated with `\n`. |

### Forcible

lingress provides the possibility for some settings (where it makes sense) to force those settings. They are marked at [section parameters](#parameters) in the column _Forcible_. A forced value is prefixed by a `!`; see below:

Usually the upstream's behaviors always wins. Assuming the following settings:
* lingress has flag `--response.headers=X-Foo:fromLingress`
* Ingress configuration has annotation `lingress.echocat.org/headers.response: "X-Foo:fromIngressConfig"`
* The upstream sends the header `X-Foo:fromUpstream`
  ... in this case the client will receive `X-Foo:fromUpstream`.

If now the ingress configuration has annotation `lingress.echocat.org/headers.response: "!X-Foo:fromIngressConfig"`<br>
... the client will receive `X-Foo:fromIngressConfig` whatever the upstream will send.

If not lingress has flag `--response.headers=!X-Foo:fromLingress`<br>
... the client will receive `X-Foo:fromLingress`.

This means: The prefix `!` will force this value over the lower ones. If the ingress configuration defines it, but not lingress, the ingress configuration will always win. If lingress defines `!`, it will always win, regardless what the ingress configuration and upstream defines.

As a result, this feature helps:
1. Force upstream applications to always have headers/behaviors set, although they are not supported by them itself.
2. Ensure the whole cluster always follow general security rules and no other ingress configuration or upstream can change it.

In the [section configuration parameters](#parameters) `L` means supported by lingress and `C` means supported by the ingress configuration.

## Helm values

Please refer all supported Helm values and their documentation: [charts/lingress/values.yaml](../charts/lingress/values.yaml)

## More topics
* [Setup](setup.md)
* [Headers](headers.md)
* [Examples](examples.md)
* [Overview](../README.md)
