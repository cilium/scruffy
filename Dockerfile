FROM docker.io/library/golang:1.20.4@sha256:690e4135bf2a4571a572bfd5ddfa806b1cb9c3dea0446ebadaf32bc2ea09d4f9 as builder
LABEL maintainer="maintainer@cilium.io"
ADD . /go/src/github.com/cilium/scruffy
WORKDIR /go/src/github.com/cilium/scruffy
RUN make scruffy
RUN strip scruffy

FROM docker.io/library/alpine:3.18.0@sha256:02bb6f428431fbc2809c5d1b41eab5a68350194fb508869a33cb1af4444c9b11 as certs
RUN apk --update add ca-certificates git
LABEL maintainer="maintainer@cilium.io"
COPY --from=builder /go/src/github.com/cilium/scruffy/scruffy /usr/bin/scruffy
ENTRYPOINT ["/usr/bin/scruffy"]
