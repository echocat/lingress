# lingress

lingress = The lean ingress controller.

In contrast to the most other ingress controllers for Kubernetes, this one does not come with extreme many different features and tries to be extreme dynamic. It tries to be as [lean](https://en.wikipedia.org/wiki/Lean_software_development) as possible.

## Topics

1. [Main principles](#main-principles)
2. [Configuration parameters](#configuration-parameters)
   1. [Forcible](#forcible)
3. [Headers](#headers)
4. [Contributing](#contributing)
5. [License](#license)

## Main principles

1. **Minimal configuration**: Just one single configuration and this as short as possible; without repeating common stuff for all ingress configurations (like ensure CORS). The majority of the ingress controllers coming nowadays with other many several configurations and/or you have to define different settings for http and https.<br><br>

2. **Centralize standards**: Settings like [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) or [HSTS](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) usually settings you want to have for your whole application. For the majority of the ingress controllers, these settings are made per ingress configuration. lingress provides the feature to centralize these settings and even [force them](#forcible).<br><br>

3. **HTTPS by default**: As lingress makes no difference between http and https by default, it also forces (which is nowadays industry standards) the clients to https (if the clients comes from http).<br><br>

4. **Industry standards by default**: Instead of always activating some industry standards or even need plugins to add them lingress, supports the important ones out of the box. See [headers sections](#headers) for examples. 

## Configuration parameters

| CLI Flag | Annotation | Default | [Forcible](#forcible) | Description |
|--|--|---|--|--|
| `--accessLog.queueSize` | | `5000` | | Maximum number of accessLog elements that could be queue before blocking. |
| `--accessLog.inline` | | `false` | | MInstead of exploding the accessLog entries into sub-entries everything is inlined into the root object. |
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
| `--log.format` | | `text` | | Defines in which format will be logged in. Can be `text` and `json`. |
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

lingress provides the possibility for some settings (where it makes sense) to force those settings. They are marked at [section configuration parameters](#configuration-parameters) in the column _Forcible_. A forced value is prefixed by a `!`; see below:

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

In the [section configuration parameters](#configuration-parameters) `L` means supported by lingress and `C` means supported by the ingress configuration.

## Headers

lingress provides many industry standard headers out-of-the-box, without enabling them.

| Name | Client▶️ | LoadBalancer▶️ | ▶️Upstream | ▶️Client | Description |
|--|--|--|--|--|--|
| `X-Correlation-Id` | ✅ | ✅ | ✅ | ✅ | Has to be base64 encoded UUID, without padding. This one is forwarded through the whole lifecycle of the requests, to help a client to identify its own resources at responses. If not provided, it will be generated by lingress. This is quite similar to `X-Request-Id`. |
| `X-Forwarded-For` | | ✅ | ✅ |  | <p>If lingress is behind of another load balancer and `--server.behindReverseProxy=true`, lingress will use this header to identify the remote client.</p><p>In any case this header is send to the upstream to tell the upstream the IP/host of the remote client.</p> |
| `X-Forwarded-Host` | | ✅ | ✅ |  | Same as `X-Forwarded-For` but for the host name which was requested by the client. |
| `X-Forwarded-Proto` | | ✅ | ✅ |  | Same as `X-Forwarded-For` but for the proto/scheme which was requested by the client. Can be `http` or `https`. |
| `X-Forwarded-Prefix` | | ✅ | ✅ |  | <p>If in the ingress configuration there was a [spec.rules.http.paths.path](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource) used, the matched prefix is contained in this header, send to the upstream.</p><p>If lingress is behind of another load balancer and `--server.behindReverseProxy=true`, lingress will use this header to identify the original uri if it was rewritten by the load balancer.</p> |
| `X-Original-URI` | | ✅ | | | If lingress is behind of another load balancer and `--server.behindReverseProxy=true`, lingress will use this header to identify the original uri if it was rewritten by the load balancer. This header is used in cases if the load balancer does not support `X-Forwarded-Prefix`. |
| `X-Real-IP` | | ✅ | | | If lingress is behind of another load balancer and `--server.behindReverseProxy=true`, lingress will use this header to identify the remote client. This header is used in cases if the load balancer does not support `X-Forwarded-For`. |
| `X-Reason` | | | | ✅ | This header is send to the clients in the response to tell why some actions happens. Currently supported: `cors-options`, `force-secure` and `not-whitelisted`. |
| `X-Request-Id` | | ✅ | ✅ | ✅ | Is a base64 encoded UUID, without padding. This one is forwarded through the whole lifecycle of the requests, once the request reached lingress. It is always generated by lingress and this cannot be changed. This is quite similar to `X-Correlation-Id`. |
| `X-Source` | | | | ✅ | This header is send to the clients in the response to explain from which ingress configuration the response was coming from. Absent means: No matching ingress configuration was found. Usually the fallback will answer then. |

## Contributing

lingress is an open source project by [echocat](https://echocat.org).
So if you want to make this project even better, you can contribute to this project on [GitHub](https://github.com/echocat/lingress)
by [fork us](https://github.com/echocat/lingress/fork).

If you commit code to this project, you have to accept that this code will be released under the [license](#license) of this project.

## License

See the [LICENSE](LICENSE) file.
