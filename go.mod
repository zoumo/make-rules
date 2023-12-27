module github.com/zoumo/make-rules

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/go-git/go-git/v5 v5.11.0
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/zoumo/golib v0.0.0-20220223062151-794bff922af0
	github.com/zoumo/goset v0.2.0
	go.uber.org/zap v1.19.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
)
