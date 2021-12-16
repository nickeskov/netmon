export GO111MODULE=on

SOURCE_DIRS = cmd pkg

.PHONY: vendor vetcheck fmtcheck clean build gotest mod-clean

all: vendor vetcheck fmtcheck build gotest mod-clean

vendor:
	go mod vendor

vetcheck:
	go vet ./...
	golangci-lint run

fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

clean:
	rm -rf build/

build:
	go build -o build/netmon ./cmd/

gotest:
	go test -cover -race -covermode=atomic ./...

mod-clean:
	go mod tidy


mock:
	@mockgen -source=pkg/monitor/scraper.go -destination=pkg/monitor/scraper_mock.go -package=monitor
	@mockgen -source=pkg/monitor/netmon.go -destination=pkg/monitor/netmon_mock.go -package=monitor
