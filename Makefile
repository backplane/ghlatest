.PHONY: build lint clean test

build: lint test ghlatest

ghlatest: *.go */*.go
	@echo '==> Building $@'
	go build -o "$@"

lint:
	@echo '==> Linting'
	go fmt
	go vet
	staticcheck

# test:
# 	@echo '==> Testing'
# 	go test -v

clean:
	@echo '==> Cleaning'
	rm -rf -- ghlatest test
