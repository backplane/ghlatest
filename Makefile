.PHONY: build lint clean test

build: lint test ghlatest

ghlatest: *.go
	@echo '==> Building $@'
	go build -o "$@" $^

lint: *.go
	@echo '==> Linting'
	go fmt
	go vet
	staticcheck

# test:
# 	@echo '==> Testing'
# 	go test -v

clean:
	@echo '==> Cleaning'
	rm -rf -- ghlatest
