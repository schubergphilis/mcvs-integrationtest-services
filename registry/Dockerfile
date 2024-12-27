FROM registry:3.0.0-rc.2
ARG APPLICATION
RUN apk update && \
    apk upgrade
# A custom config file is needed to disable default cache
# By default the cache is "blobdescriptor: inmemory"
# The config-example.yml file of this repo can be found at:
# https://github.com/distribution/distribution-library-image
# File is copied to two different locations to keep
# compatibility with registry:2.x and registry:3.x
COPY ./config-example.yml /etc/docker/registry/config.yml
COPY ./config-example.yml /etc/distribution/config.yml
COPY ./backup-registry /var/lib/registry
