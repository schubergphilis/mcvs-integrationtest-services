FROM golang:1.23.3-alpine AS builder
ENV USERNAME=mcvs-integrationtest-services
ENV HOME=/home/${USERNAME}
RUN adduser -D -g '' ${USERNAME}
COPY . /go/${USERNAME}/
WORKDIR /go/${USERNAME}/cmd/${USERNAME}
RUN apk add --no-cache \
        curl=~8 \
        git=~2 && \
    CGO_ENABLED=0 go build -buildvcs=false && \
    find ${HOME}/ -mindepth 1 -delete && \
    chown 1000 -R ${HOME} && \
    chmod 0700 -R ${HOME}

FROM alpine:3.20.3
ENV USERNAME=mcvs-integrationtest-services
ENV HOME=/home/${USERNAME}
ENV PATH=${HOME}/bin:${PATH}
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/${USERNAME}/cmd/${USERNAME}/${USERNAME} /usr/local/bin/${USERNAME}
COPY --from=builder /home/${USERNAME} ${HOME}/
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        curl=~8 \
        libcrypto3=~3 \
        libssl3=~3 && \
    chown 1000 -R ${HOME} && \
    chmod 0700 -R ${HOME} && \
    rm -rf /var/cache/apk/*
VOLUME ["/tmp","/home/${USERNAME}"]
USER ${USERNAME}
EXPOSE 1323
ENTRYPOINT ["mcvs-integrationtest-services"]
