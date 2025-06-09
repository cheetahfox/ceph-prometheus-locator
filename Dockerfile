FROM golang:alpine3.20 AS builder

RUN apk add --no-cache gcc git musl-dev build-base openssh-client

RUN mkdir /go/src/github.com
RUN mkdir /go/src/github.com/cheetahfox
RUN mkdir /go/src/github.com/cheetahfox/ceph-prometheus-locator


COPY ./ /go/src/github.com/cheetahfox/ceph-prometheus-locator/

WORKDIR /go/src/github.com/cheetahfox/ceph-prometheus-locator

RUN go build

FROM alpine:3.20.2 
COPY --from=builder /go/src/github.com/cheetahfox/ceph-prometheus-locator/ceph-prometheus-locator /ceph-prometheus-locator
EXPOSE 8080
CMD ["/ceph-prometheus-locator"]