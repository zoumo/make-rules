module github.com/zoumo/make-rules

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-logr/logr v0.2.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/zoumo/goset v0.2.0
	k8s.io/klog/v2 v2.2.0
)

replace k8s.io/klog/v2 => k8s.io/klog/v2 v2.2.0
