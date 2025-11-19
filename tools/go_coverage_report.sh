mkdir -p test/coverage
go test -race -v $(go list ./...) -coverprofile=test/coverage/coverage.out
exit 0