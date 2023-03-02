# Build CLI executable

## Requirements

1. `golang` 1.19
2. `make`

## Check formatting and run tests

Use `make check test` to run the default `gofmt` formatting and unit tests

```
$ make check test

Checking for syntax errors
gofmt -e . > /dev/null
cd clients/omnitruck; go test
PASS
ok      github.com/chef/omnitruck-service/clients/omnitruck     0.396s
```

## Full build with swagger docs and cli executable

To generate the latest swagger API docs and compile the CLI executable run `make all`. Final executable will be available as `bin/omnitruck-service`

```bash
$ make all

swag init -o docs -d services --parseDependency --instanceName OmnitruckApi
2023/02/22 14:24:37 Generate swagger docs....
2023/02/22 14:24:38 Generate general API Info, search dir:services
2023/02/22 14:24:38 Generating omnitruck.ItemList
2023/02/22 14:24:38 Generating services.ErrorResponse
2023/02/22 14:24:38 Generating omnitruck.PlatformList
2023/02/22 14:24:38 Generating omnitruck.PackageList
2023/02/22 14:24:38 Generating omnitruck.PlatformVersionList
2023/02/22 14:24:38 Generating omnitruck.ArchList
2023/02/22 14:24:38 Generating omnitruck.PackageMetadata
2023/02/22 14:24:38 create docs.go at  docs/OmnitruckApi_docs.go
2023/02/22 14:24:38 create swagger.json at  docs/OmnitruckApi_swagger.json
2023/02/22 14:24:38 create swagger.yaml at  docs/OmnitruckApi_swagger.yaml
Building cli
go build -o bin/
```

## Only build CLI

Use `make build` to build just the CLI executable

```bash
$ make build
Building cli
go build -o bin/
```