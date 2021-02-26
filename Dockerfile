FROM docker.io/library/golang:1.17.2@sha256:cf615c1499e8bc2cf00696ba234cddd47fdb8d9a3b37b7c35726e46ee4ae08cc as builder
LABEL maintainer="maintainer@cilium.io"
ADD . /go/src/github.com/cilium/scruffy
WORKDIR /go/src/github.com/cilium/scruffy
RUN make scruffy
RUN strip scruffy

FROM docker.io/library/alpine:3.14.2@sha256:e1c082e3d3c45cccac829840a25941e679c25d438cc8412c2fa221cf1a824e6a as certs
RUN apk --update add ca-certificates git
LABEL maintainer="maintainer@cilium.io"
COPY --from=builder /go/src/github.com/cilium/scruffy/scruffy /usr/bin/scruffy
ENTRYPOINT ["/usr/bin/scruffy"]
