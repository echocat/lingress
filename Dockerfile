FROM golang:1.12 as builder
ARG GROUP_ID="echocat"
ARG ARTIFACT_ID="lingress"
ARG REVISION="latest"
ARG BRANCH="development"
ARG SSH_PRIVATE_KEY
COPY  .  /src
WORKDIR  /src
RUN export GOPATH=/src/.go && mkdir -p $GOPATH \
    && go mod download \
    && go run github.com/gobuffalo/packr/packr clean \
    && go run github.com/gobuffalo/packr/packr -z -i fallback \
    && CGO_ENABLED=0 go run github.com/echocat/lingress/build -o /tmp/lingress ./lingress

FROM scratch
USER 5000
WORKDIR /
COPY --from=builder /tmp/lingress /lingress
ENTRYPOINT ["/lingress"]
