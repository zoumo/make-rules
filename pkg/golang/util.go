package golang

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	MinimumGoVersion = "1.13.0"
)

func VerifyGoVersion() error {
	v := strings.TrimLeft(runtime.Version(), "go")
	sv1, err := semver.NewVersion(v)
	if err != nil {
		return err
	}

	sv2, err := semver.NewVersion(MinimumGoVersion)
	if err != nil {
		return err
	}

	if sv1.Compare(sv2) < 0 {
		return fmt.Errorf(`detected go version: %v. This project requires %v or greater. Please install %v or later`, v, MinimumGoVersion, MinimumGoVersion)
	}

	return nil
}
