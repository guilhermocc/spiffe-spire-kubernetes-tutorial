.PHONY: docker-build
docker-build: greeter-server-image greeter-client-image

.PHONY: greeter-server-image
greeter-server-image:
	docker build --target greeter-server -t greeter-server:demo .

.PHONY: greeter-client-image
greeter-client-image:
	docker build --target greeter-client -t greeter-client:demo .

.PHONY: deploy
deploy: docker-build
	kind load docker-image greeter-server:demo --name "spire-example"
	kind load docker-image greeter-client:demo --name "spire-example"
