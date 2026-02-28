ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-trixie AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG SOURCE=https://github.com/Kubedoll-Heavy-Industries/mcp-helm
ARG TARGETARCH=amd64

RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
  go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o /out/mcp-helm ./cmd/mcp-helm

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG SOURCE=https://github.com/Kubedoll-Heavy-Industries/mcp-helm

LABEL org.opencontainers.image.title="mcp-helm" \
  org.opencontainers.image.description="MCP server for interacting with Helm repositories and charts" \
  org.opencontainers.image.licenses="MIT" \
  org.opencontainers.image.source=$SOURCE \
  org.opencontainers.image.version=$VERSION \
  org.opencontainers.image.revision=$COMMIT \
  org.opencontainers.image.created=$DATE

EXPOSE 8012

COPY --from=build /out/mcp-helm /mcp-helm

ENTRYPOINT ["/mcp-helm"]
CMD ["--listen=:8012", "--transport=http"]

FROM alpine:3.23.3 AS debug

RUN apk add --no-cache ca-certificates tzdata curl

EXPOSE 8012

COPY --from=build /out/mcp-helm /mcp-helm

ENTRYPOINT ["/mcp-helm"]
CMD ["--listen=:8012", "--transport=http"]
