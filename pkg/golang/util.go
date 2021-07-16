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

func VerifyGoVersion(minimumGoVersion string) error {
	v := strings.TrimPrefix(runtime.Version(), "go")
	sv1, err := semver.NewVersion(v)
	if err != nil {
		return err
	}

	if len(minimumGoVersion) == 0 {
		minimumGoVersion = MinimumGoVersion
	}

	mv := strings.TrimPrefix(minimumGoVersion, "go")
	sv2, err := semver.NewVersion(mv)
	if err != nil {
		return err
	}

	if sv1.Compare(sv2) < 0 {
		return fmt.Errorf(`detected go version: %v. This project requires %v or greater. Please install %v or later`, v, minimumGoVersion, minimumGoVersion)
	}

	return nil
}
