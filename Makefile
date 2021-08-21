#----TEST----

test-install-gotest:
	go get -u github.com/rakyll/gotest

test:
	gotest -v || go clean -testcache
	go clean -testcache


#----LINT----

lint-install:
	# binary will be $(go env GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.41.1
	golangci-lint --version

lint:
	@if [ -z `which golangci-lint 2> /dev/null` ]; then \
			echo "Need to install golangci-lint, execute \"make lint-install\"";\
			exit 1;\
	fi
	golangci-lint run

lint-fix:
	golangci-lint run --fix
