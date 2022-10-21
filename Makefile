all: swagger build
swagger:
	swag init -o docs/opensource -d services/opensource --parseDependency

build:
	go build -o bin/


start:
	bin/omnitruck-service start