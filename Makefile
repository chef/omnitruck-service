TESTS = test_omnitruck_client test_services

BINARY_NAME ?= omnitruck-service
BUILD_ARCH ?= amd64
BUILD_OS ?= linux
BUILD_PROXY ?= https://proxy.golang.org,direct

all: swagger build
swagger:
	swag init -g httpserver/routes.go -d .,internal/api/handler --output ./docs --parseDependency --parseInternal --instanceName OmnitruckApi
	swagger2openapi docs/OmnitruckApi_swagger.json -o docs/OmnitruckApi_openapi3.json 

build:
	@echo "Building cli"
	GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) GOPROXY=$(BUILD_PROXY) go build -o bin/$(BINARY_NAME)

test: 
	go test -race -vet=off ./...

test_omnitruck_client:
	cd clients/omnitruck; go test -v
test_services: 
	cd services; go test -v

image: 
	@echo "Creating docker image"
	docker build -t chef/omnitruck-services:dev .

image_linux:
	@echo "Creating docker image for linux"
	docker buildx build --platform=linux/amd64 -t chef/omnitruck-services:latest .

image_push: image_linux 
	@echo "Pushing image to ${DOCKER_HUB}"
	docker tag chef/omnitruck-services:latest ${DOCKER_HUB}
	docker push ${DOCKER_HUB}

docker_run: image
	@echo "Running docker image chef/omnitruck-services"
	docker run -d --rm -p 3000:3000 -p 3001:3001 -p 3002:3002 chef/omnitruck-services:dev

start:
	bin/$(BINARY_NAME) start

check:
	@echo "Checking for syntax errors"
	gofmt -e . > /dev/null

clean:
	rm bin/$(BINARY_NAME)