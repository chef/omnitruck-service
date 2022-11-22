TESTS = test_omnitruck_client

all: swagger build
swagger:
	swag init -o docs -d services --parseDependency --instanceName OmnitruckApi

build:
	@echo "Building cli"
	go build -o bin/

test: $(TESTS)

test_omnitruck_client:
	cd clients/omnitruck; go test

image: 
	@echo "Creating docker image"
	docker build -t chef/omnitruck-services:dev .

docker_run:
	@echo "Running docker image chef/omnitruck-services"
	docker run -d --rm -p 3000:3000 -p 3001:3001 -p 3002:3002 chef/omnitruck-services:dev

start:
	bin/omnitruck-service start

check:
	@echo "Checking for syntax errors"
	gofmt -e . > /dev/null

clean:
	rm bin/omnitruck-service