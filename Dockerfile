FROM alpine:latest

ARG  version
ARG  image
COPY dist/lingress-linux-amd64 /usr/bin/lingress

RUN chmod +x /usr/bin/lingress \
    && apk add --no-cache curl \
    && rm -rf \
        /usr/share/man \
        /tmp/* \
        /var/cache/apk \
    && mkdir -p /usr/lib/lingress \
    && echo -n "${version}" > /usr/lib/lingress/version \
    && echo -n "${image}" > /usr/lib/lingress/image

ENTRYPOINT ["/usr/bin/lingress"]
