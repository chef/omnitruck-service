go test -race -v $(go list ./...) -coverprofile=cov.out
exit 0