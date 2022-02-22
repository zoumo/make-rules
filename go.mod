module github.com/zoumo/make-rules

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-logr/logr v1.2.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/zoumo/golib v0.0.0-20220222092837-7e3a4fe9a2f0
	github.com/zoumo/golog v0.4.1
	github.com/zoumo/goset v0.2.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	github.com/zoumo/golog => github.com/zoumo/golog v0.4.1
)
