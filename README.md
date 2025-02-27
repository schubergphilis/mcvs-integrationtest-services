# mcvs-integrationtest-services

This repository, "mcvs-integrationtest-services" provides a versatile Docker
image designed to mimic multiple services, including Okta, AWS, and others. The
primary purpose of this Docker image is to facilitate comprehensive testing
environments where developers can simulate real-world scenarios involving
various services without needing to interact with the actual external services.
This approach is especially beneficial in integration, component and
end-to-end (e2e) testing, ensuring that all aspects of the application's
interaction with these services are thoroughly vetted.

In conjunction with the [dockertest](https://github.com/ory/dockertest) library,
this image allows developers to write robust and extensive tests that cover a
wide range of scenarios. Dockertest is a Go package essential for running Docker
containers as part of the testing process. By integrating these simulated
services, developers can streamline their testing process, detect potential
issues early, and maintain the stability and reliability of the system. This
repository thus plays a crucial role in enhancing the overall quality and
security of the application by ensuring that it performs as expected in various
integrated environments.

Note: This image can be used with other programming languages as well, as long
as they have a framework similar to go-dockertest

## MCVS-IntegrationTest-Service

### Build

```zsh
docker build -t mcvs-integrationtest-services .
```

### Run

```zsh
docker run -p 9999:1323 -it mcvs-integrationtest-services
```

### Test

```zsh
curl \
  -X POST http://localhost:9999/authorization/users \
  -H "Content-Type: application/json" \
  -d '{"action":"listLabels","email":"something@example.com","facility":"a","group":"a","name":"someName"}'
```

## MCVS-Stub-Server

A simple HTTP server which can configure endpoints with a given response. This
can be used as a stub server to mimic behavior of other services.

### Build

```zsh
docker build -t stub-server --build-arg APPLICATION=mcvs-stub-server .
```
**Note:**
When building locally, the tzdata package download might fail. To fix this on your local machine, change the version in [Dockerfile:17](Dockerfile)

### Run

```zsh
docker run -p 8080:8080 stub-server
```

### Test

**Configuring**

```
curl --location 'localhost:8080/configure' \
--header 'Content-Type: application/json' \
--data '{
    "path": "/foo",
    "response": {"foo": "bar"}
}'
```

**Hit a configured endpoint**

```
curl --location 'localhost:8080/foo'
```

**Reset a configured endpoint**

```
curl --location 'localhost:8080/reset' \
--header 'Content-Type: application/json' \
--data '{
    "path": "/foo"
}'
```

## Okta

Generate a valid Okta JSON Web Token (JWT).

### Build

```zsh
docker build -t oktamock --build-arg APPLICATION=oktamock .
```

### Run

```zsh
docker run -p 8080:8080 oktamock
```

### Test

```zsh
curl http://localhost:8080/token
```
