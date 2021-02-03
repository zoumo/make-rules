package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	pretty = false
)

// Info contains versioning information.
// TODO: Add []string of api versions supported? It's still unclear
// how we'll want to distribute that information.
type Info struct {
	GitCommit    string `json:"gitCommit"`
	GitRemote    string `json:"gitRemote"`
	GitVersion   string `json:"gitVersion"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// Pretty returns a pretty output representation of Info
func (info Info) JSON() string {
	str, _ := json.Marshal(info)
	return string(str)
}

// Pretty returns a pretty output representation of Info
func (info Info) PrettyJSON() string {
	str, _ := json.MarshalIndent(info, "", "    ")
	return string(str)
}

// String returns info as a human-friendly version string.
func (info Info) String() string {
	return info.GitVersion
}

// IsDirty returns true if the project is compiled in a dirty git tree or without a valid version
func (info Info) IsDirty() bool {
	if info.GitVersion == "" ||
		info.GitVersion == "unknown" ||
		strings.Contains(info.GitVersion, "-dirty") {
		return true
	}
	return false
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Info {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in pkg/version/base.go
	return Info{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GitRemote:    gitRemote,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// NewCommand returns a command to show the version
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			out := Get().String()
			if pretty {
				out = Get().PrettyJSON()
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.BoolVar(&pretty, "json", pretty, "show version as pretty json output")
	return cmd
}
