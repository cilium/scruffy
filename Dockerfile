FROM docker.io/library/golang:1.18.4@sha256:9349ed889adb906efa5ebc06485fe1b6a12fb265a01c9266a137bb1352565560 as builder
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
