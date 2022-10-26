all: swagger build
swagger:
	swag init -o docs/opensource -d services/opensource --parseDependency --instanceName Opensource
	swag init -o docs/trial -d services/trial --parseDependency --instanceName Trial

build:
	go build -o bin/


start:
	bin/omnitruck-service start
