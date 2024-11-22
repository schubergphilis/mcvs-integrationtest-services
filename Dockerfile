FROM golang:1.23.3-alpine as builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
COPY . .

ARG APPLICATION=mcvs-integrationtest-services
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main ./cmd/$APPLICATION

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main ./main
CMD ["./main"]
