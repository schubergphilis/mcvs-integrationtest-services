# mcvs-integrationtest-services

## Build

```zsh
docker build -t mcvs-integrationtest-services .
```

## Run

```zsh
docker run -p 9999:1323 -it mcvs-integrationtest-services
```

## Test

```zsh
curl \
  -X POST http://localhost:9999/authorization/users \
  -H "Content-Type: application/json" \
  -d '{"action":"listLabels","email":"something@example.com","facility":"a","group":"a","name":"someName"}'
```
