all: swagger build
swagger:
	swag init -o docs/opensource -d services/opensource --parseDependency --instanceName Opensource
	swag init -o docs/trial -d services/trial --parseDependency --instanceName Trial
	swag init -o docs/commercial -d services/commercial --parseDependency --instanceName Commercial

build:
	go build -o bin/


start:
	bin/omnitruck-service start
