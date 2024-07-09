
HOSTNAME=localhost

.PHONY: generate clean
generate: ca server client

clean:
	rm ca -r || true
	rm server -r || true
	rm client -r || true

ca:
	cfssl genkey -initca cacsr.json | cfssljson -bare ca
	mkdir ca
	mv ca.csr ca/
	mv ca.pem ca/
	mv ca-key.pem ca/

server: ca
	cfssl gencert -ca ca/ca.pem -ca-key ca/ca-key.pem -profile server -hostname=$(HOSTNAME) certcsr.json | cfssljson -bare server
	mkdir server
	mv server.csr server/
	mv server.pem server/
	mv server-key.pem server/

client: ca
	cfssl gencert -ca ca/ca.pem -ca-key ca/ca-key.pem -profile client -hostname=$(HOSTNAME) certcsr.json | cfssljson -bare client
	mkdir client
	mv client.csr client/
	mv client.pem client/
	mv client-key.pem client/

.PHONY: more-client more-server
more-client: ca
	rm client -r || true
	make client

more-server: ca
	rm server -r || true
	make server