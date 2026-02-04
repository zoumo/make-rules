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

// AddGlobalFlags registers global flags
func AddGlobalFlags(fs *pflag.FlagSet) {
	consolog.InitFlags(fs)
}

// normalize replaces underscores with hyphens
func normalize(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}
