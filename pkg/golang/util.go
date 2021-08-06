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

func getGoVersion() (string, error) {
	cmd := runner.NewRunner("go")
	out, err := cmd.RunCombinedOutput("version")
	if err != nil {
		return "", err
	}
	output := strings.Split(string(out), " ")

	if len(output) != 4 {
		return "", fmt.Errorf("get error go version output: %v", string(out))
	}
	return output[2], nil
}

func VerifyGoVersion(minimumGoVersion string) error {
	version, err := getGoVersion()
	if err != nil {
		return err
	}
	v := strings.TrimPrefix(version, "go")
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
