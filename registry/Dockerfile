FROM registry:3.1.2@sha256:ddf754342cfc8acc51a56d5d0ab6af06826461864460636d8bd5c546dab2a7b8 AS registry
ARG APPLICATION
RUN apk update && \
    apk upgrade && \
    adduser -D -u 1000 ${APPLICATION}
# A custom config file is needed to disable default cache
# By default the cache is "blobdescriptor: inmemory"
# The config-example.yml file of this repo can be found at:
# https://github.com/distribution/distribution-library-image
# File is copied to two different locations to keep
# compatibility with registry:2.x and registry:3.x
COPY config-example.yml /etc/docker/registry/config.yml
COPY config-example.yml /etc/distribution/config.yml
COPY backup-registry /var/lib/registry

# Copying the filesystem into a single layer prevents the packages that
# apk upgrade replaces from being stored twice, which Dive fails on. The
# settings below repeat those of the registry base image.
FROM scratch
ARG APPLICATION
COPY --from=registry / /
ENV OTEL_TRACES_EXPORTER=none \
    PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
VOLUME ["/var/lib/registry"]
EXPOSE 5000
USER ${APPLICATION}
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/etc/distribution/config.yml"]
