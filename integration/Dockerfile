# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25
ARG HELM_VERSION=v3.18.4
ARG NODE_VERSION=22.18.0

FROM golang:${GO_VERSION}-bookworm AS base
ARG HELM_VERSION
ARG TARGETARCH
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && curl -fsSL -o /tmp/helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" \
    && tar -xzf /tmp/helm.tar.gz -C /tmp \
    && mv "/tmp/linux-${TARGETARCH}/helm" /usr/local/bin/helm \
    && rm -rf /tmp/helm.tar.gz "/tmp/linux-${TARGETARCH}" /var/lib/apt/lists/*

COPY --chmod=755 integration/container-run.sh /usr/local/bin/integration-run

FROM node:${NODE_VERSION}-bookworm AS node-runtime

FROM base AS upstream
# cdk8splus34's upstream Go bindings predate constructs' mixin API. Keep the
# importer on the corresponding cdk8s CLI line so CRD bindings compile with it.
ARG CDK8S_CLI_VERSION=2.200.156
COPY --from=node-runtime /usr/local/ /usr/local/
RUN npm install --global "cdk8s-cli@${CDK8S_CLI_VERSION}" "constructs@10.4.2"
ENV CDK8S_IMPLEMENTATION=upstream
ENTRYPOINT ["/usr/local/bin/integration-run"]

FROM base AS pure
WORKDIR /purecdk8s
COPY . .
RUN go build -o /usr/local/bin/purecdk8s ./cmd/purecdk8s
COPY --chmod=755 integration/migrate.sh /usr/local/bin/migrate-purecdk8s
ENV CDK8S_IMPLEMENTATION=pure
ENTRYPOINT ["/usr/local/bin/integration-run"]
