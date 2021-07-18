.PHONY: godoc
godoc:
	godoc -http :8080

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
