TESTS = test_omnitruck_client test_services

all: deps swagger build
deps: 
	npm install -g swagger2openapi
swagger:
		swag init -g services/main.go --output ./docs --parseDependency --parseInternal --instanceName OmnitruckApi
		swagger2openapi docs/OmnitruckApi_swagger.json -o docs/OmnitruckApi_openapi3.json

build:
	@echo "Building cli"
	go build -o bin/

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
	bin/omnitruck-service start

check:
	@echo "Checking for syntax errors"
	gofmt -e . > /dev/null

clean:
	rm bin/omnitruck-service