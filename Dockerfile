FROM golang:1.10-alpine@sha256:9de80ce12179b571ed46f9d1fb12261640a6d5af04689d6536a07f9dc23eae50 as alpine-build

RUN apk add --no-cache \
    make \
    git \
    upx

WORKDIR /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/
COPY . /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/

RUN make build-alpine

FROM golang:1.10@sha256:2ffa2f093d20c46e86435626f11bf163797400cf8f7cf14ecdc6403f1930045c as build

RUN apt-get update && \
    apt-get install -y \
        upx-ucl

WORKDIR /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/
COPY . /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/

RUN make build-all test-unit

FROM scratch as artifacts

COPY --from=alpine-build /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/bin/* /
COPY --from=build /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/bin/* /

FROM vault:latest

COPY --from=artifacts /vault-circleci-auth-plugin_

ENV VAULT_ADDR=http://127.0.0.1:8200

ENV VAULT_TOKEN=root

RUN mkdir /vault/plugins

COPY launch.sh /launch.sh

COPY --from=build /go/src/github.com/marcboudreau/vault-circleci-auth-plugin/vault-circleci-auth-plugin /vault/plugins/

RUN chown vault:vault /vault/plugins/vault-circleci-auth-plugin

CMD [ "/launch.sh" ]