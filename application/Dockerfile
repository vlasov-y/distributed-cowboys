# syntax=docker/dockerfile:1
#  ┌─┐┌─┐┬  ┬─┐┌┐┐┌─┐
#  │ ┬│ ││  │─┤││││ ┬
#  ┘─┘┘─┘┘─┘┘ ┘┘└┘┘─┘

FROM --platform=$BUILDPLATFORM golang:1.21.0-alpine3.18 AS builder
WORKDIR /app

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ENV CGO_ENABLED=0
ENV GOOS ${TARGETOS:-linux}
ENV GOARCH ${TARGETARCH}

COPY go.mod go.sum ./
RUN sha256sum go.mod go.sum > go.lock
COPY *.go ./
RUN if ! sha256sum -c go.lock; then \
      go mod vendor && \
      sha256sum go.mod go.sum > go.lock; \
    fi
RUN go build -a \
      -o build \
      -ldflags="-s -w" \
      -trimpath \
      -buildmode=exe

#  ┬─┐┬ ┐┌┐┐┌┐┐┬─┐┬─┐
#  │┬┘│ │││││││├─ │┬┘
#  ┘└┘┘─┘┘└┘┘└┘┴─┘┘└┘

FROM scratch AS runner
WORKDIR /app
ENTRYPOINT [ "/app/build" ]
COPY --chown=1000:1000 --from=builder /app/build ./build

