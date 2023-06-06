ARG GO_VERSION=1.20.4
ARG ALPINE_VERSION=3.18
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ENV GOBIN=/usr/local/bin
RUN mkdir /src
COPY . /src/

WORKDIR /src/envoy-to-harald
RUN pwd && ls -lah / && ls -lah /src
RUN go install .

WORKDIR /src/test-app
RUN go install .

FROM ghcr.io/maxmoehl/harald:main

RUN apk add --no-cache yq curl

COPY --from=builder /usr/local/bin/envoy-to-harald /usr/local/bin/test-app /usr/local/bin/
COPY run.sh /usr/local/bin/run.sh

ENTRYPOINT [ "/usr/local/bin/run.sh" ]
