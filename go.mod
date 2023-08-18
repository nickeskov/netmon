module github.com/nickeskov/netmon

go 1.18

// exclude unused vulnerable dependencies
exclude (
	golang.org/x/text v0.3.0
	golang.org/x/text v0.3.3
)

require (
	github.com/gammazero/deque v0.2.1
	github.com/golang/mock v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.25.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
