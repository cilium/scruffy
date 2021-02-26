all: local

docker-image:
	docker build -t quay.io/cilium/scruffy:${VERSION} .

tests:
	go test -mod=vendor ./...

scruffy: tests
	CGO_ENABLED=0 go build -mod=vendor -a -installsuffix cgo -o $@ ./cmd/main.go

local: scruffy
	strip scruffy

clean:
	rm -fr scruffy
