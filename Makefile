build:
	swag init -o docs/opensource -d services/opensource --parseDependency
	go build -o bin/


start:
	bin/omnitruck-service start