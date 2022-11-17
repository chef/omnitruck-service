TESTS = test_omnitruck_client

all: swagger build
swagger:
	swag init -o docs -d services --parseDependency --instanceName OmnitruckApi

build:
	go build -o bin/

test: $(TESTS)

test_omnitruck_client:
	cd clients/omnitruck; go test

start:
	bin/omnitruck-service start
