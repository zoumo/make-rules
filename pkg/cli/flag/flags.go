package flag

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/zoumo/golib/log/consolog"
)

// WordSepNormalizeFunc changes all flags that contain "_" separators
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(normalize(name))
	}
	return pflag.NormalizedName(name)
}

// AddGlobalFlags explicitly registers flags that libraries (klog, verflag, etc.) register
// against the global flagsets from "flag" and "k8s.io/klog/v2".
// We do this in order to prevent unwanted flags from leaking into the component's flagset.
func AddGlobalFlags(fs *pflag.FlagSet) {
	// addKlogFlags(fs)
	addConsologFlags(fs)
}

func addConsologFlags(fs *pflag.FlagSet) {
	consolog.InitFlags(fs)
}

// addKlogFlags adds flags from k8s.io/klog
// func addKlogFlags(fs *pflag.FlagSet) {
// 	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
// 	klog.InitFlags(local)
// 	local.VisitAll(func(fl *flag.Flag) {
// 		fl.Name = normalize(fl.Name)
// 		fs.AddGoFlag(fl)
// 	})
// }

// normalize replaces underscores with hyphens
// we should always use hyphens instead of underscores when registering component flags
func normalize(s string) string {
	return strings.Replace(s, "_", "-", -1)
}
