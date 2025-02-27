ARG APPLICATION=mcvs-integrationtest-services

FROM golang:1.23.6-alpine AS builder
ARG APPLICATION
ENV CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux
WORKDIR /app
# By installing OS packages and updates in a separate Docker layer, developers
# can prevent redundant execution during each Docker image build, significantly
# accelerating the development process.
RUN apk update && \
    apk upgrade && \
    apk add \
      --no-cache \
        git=~2 \
        tzdata=~2025 && \
    update-ca-certificates
# By copying go.mod and go.sum files in a separate step and running go mod
# download right after, the Dockerfile ensures that dependency resolution and
# downloading are cached, thereby avoiding repeated downloads and expediting
# subsequent builds if only source code changes occur.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build \
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
