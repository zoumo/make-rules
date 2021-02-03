package version

// Base Version information.
//
// This is the fallback data used when Version information from git is not
// provided via go ldflags. It provides an approximation of the Kubernetes
// Version for ad-hoc builds (e.g. `go build`) that cannot get the Version
// information from git.
//
// If you are looking at these fields in the git tree, they look
// strange. They are modified on the fly by the build process. The
// in-tree values are dummy values used for "git archive", which also
// works for GitHub tar downloads.
//
// When releasing a new Kubernetes Version, this file is updated by
// build/mark_new_version.sh to reflect the new Version, and then a
// git annotated tag (using format vX.Y where X == Major Version and Y
// == Minor Version) is created to point to the commit that updates
// pkg/Version/base.go
var (
	gitVersion   = "unknown"
	gitCommit    = "$Format:%H$" // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = ""            // state of git tree, either "clean" or "dirty"
	gitRemote    = ""

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
