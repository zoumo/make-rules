package golang

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/zoumo/make-rules/pkg/runner"
)

const (
	MinimumGoVersion = "1.13.0"
)

var (
	go1160 = semver.MustParse("v1.16.0")
)

func getGoVersion() (*semver.Version, error) {
	cmd := runner.NewRunner("go")
	out, err := cmd.RunCombinedOutput("version")
	if err != nil {
		return nil, err
	}
	output := strings.Split(string(out), " ")

	if len(output) != 4 {
		return nil, fmt.Errorf("get error go version output: %v", string(out))
	}
	sv, err := semver.NewVersion(strings.TrimPrefix(output[2], "go"))
	if err != nil {
		return nil, err
	}
	return sv, nil
}

func VerifyGoVersion(minimumGoVersion string) error {
	sv1, err := getGoVersion()
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
		return fmt.Errorf(`detected go version: %v. This project requires %v or greater. Please install %v or later`, sv1.String(), minimumGoVersion, minimumGoVersion)
	}

	return nil
}
