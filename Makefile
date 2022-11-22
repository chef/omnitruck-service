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

start:
	bin/omnitruck-service start

check:
	@echo "Checking for syntax errors"
	gofmt -e . > /dev/null
