# harald-cf-test

A showcase to deploy [harald](https://github.com/maxmoehl/harald) on Cloud Foundry as a docker image.

## Pushing the App

```shell
cf push harald-test -o ghcr.io/maxmoehl/harald-cf-test:latest -m 128M -k 256M
```

## Build & Push the Image

### Hosts running on `linux/amd64`

```shell
docker build -t ghcr.io/maxmoehl/harald-cf-test:latest --push .
```

### Hosts running on other architecture

Needs to be cross-compiled since Cloud Foundry only supports `amd64` containers.

First setup `buildx` following the [docker guide](https://docs.docker.com/build/building/multi-platform/), then build the image:
```shell
docker buildx build --platform linux/amd64 -t ghcr.io/maxmoehl/harald-cf-test:latest --provenance=false --push .
```
