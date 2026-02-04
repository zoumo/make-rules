module github.com/zoumo/make-rules

go 1.23.0

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/go-git/go-git/v5 v5.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.10
	github.com/zoumo/golib v0.2.2
	github.com/zoumo/goset v0.2.0
	sigs.k8s.io/yaml v1.2.0
)

require (
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.0.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.10
	k8s.io/client-go => k8s.io/client-go v0.18.10
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.130.1
)
