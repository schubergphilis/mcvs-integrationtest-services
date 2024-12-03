ARG APPLICATION=mcvs-integrationtest-services

FROM golang:1.23.3-alpine AS builder
ENV CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux
ARG APPLICATION
WORKDIR /app
COPY . .
RUN apk update && \
    apk add \
      --no-cache \
        ca-certificates=~20240705-r0 \
        git=~2 \
        tzdata=~2024 && \
    update-ca-certificates && \
    go mod download && \
    go build \
      -a \
      -installsuffix cgo \
      -o main \
      ./cmd/${APPLICATION} && \
    adduser -D -u 1000 ${APPLICATION}

FROM scratch
ARG APPLICATION
COPY --from=builder /app/main /app/main
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
USER ${APPLICATION}
ENTRYPOINT ["/app/main"]
