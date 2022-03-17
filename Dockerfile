FROM docker.io/library/golang:1.17.8-alpine3.15 as build

COPY . /build
WORKDIR /build
RUN GOOS=linux GARCH=amd64 CGO_ENABLED=0 \
    go build -o catlin ./cmd/catlin

FROM docker.io/library/alpine:3.15

RUN apk --no-cache add bash shellcheck py3-pip

RUN pip3 install pylint

WORKDIR /data

COPY --from=build /build/catlin /usr/bin/catlin

CMD [ "catlin" ]
