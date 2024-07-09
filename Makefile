.PHONY: daemon.zip
daemon.zip: daemon
	rm daemon.zip || true
	7z a daemon.zip build/daemon/


.PHONY: daemon
daemon: daemon-linux-amd64
	mkdir -p build/daemon/bin
	rm build/daemon/bin/* || true
	cp daemon-linux-amd64 build/daemon/bin/hornbill-daemon

	mkdir -p build/daemon/etc
	rm build/daemon/etc/* || true
	cp cfssl/server/* build/daemon/etc
	cp cfssl/ca/ca.pem build/daemon/etc
	cp build/daemon/wg-init.sh build/daemon/etc

.PHONY: daemon-linux-amd64
daemon-linux-amd64:
	export GOOS=linux
	export GOARCH=amd64
	go build -o daemon-linux-amd64 ./cmd/daemon/

.PHONY: api
api:
	docker build . -f build/apiserver/Dockerfile -t harbor.porama.dev/hornbill/hornbill-api

.PHONY: clean
clean:
	rm daemon-linux-amd64 daemon.zip || true