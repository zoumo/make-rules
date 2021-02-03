package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FindTargetsFrom find all dir under {workdir}/{subdir}/ with 1 depth as target
// and then filter target if it does not contains {mustContainFile}
//
// [example]:
//  workdir/
//    |- cmd
//        |- a - main.go
//        |- b - main.go
//        |- c - xxx.go
//
// call findTargetsFrom(workdir, "cmd", "main.go") -> result: ["cmd/a", "cmd/b"]
//
func FindTargetsFrom(workdir, subdir string, mustContainFile string) (targets []string, err error) {
	root := workdir + "/" + subdir + "/"
	err = filepath.Walk(root, func(fpath string, info os.FileInfo, ierr error) error {
		if ierr != nil {
			return ierr
		}
		if fpath == root {
			// continue
			return nil
		}
		if info.IsDir() {
			targets = append(targets, subdir+"/"+path.Base(fpath))
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	filtered := []string{}
	if mustContainFile != "" {
		for _, dir := range targets {
			if _, err := os.Stat(path.Join(workdir, dir, mustContainFile)); err == nil {
				filtered = append(filtered, dir)
			}
		}
	} else {
		filtered = targets
	}
	return filtered, nil
}

func FilterTargets(inputs []string, allTargets []string, prefix string) []string {
	// find build target
	if len(inputs) == 0 {
		return allTargets
	}
	filtered := []string{}
	for _, input := range inputs {
		if !strings.HasPrefix(input, prefix+"/") {
			input = prefix + "/" + input
		}
		for _, t := range allTargets {
			// check if target is valid
			if t == input {
				filtered = append(filtered, input)
			}
		}
	}
	return filtered
}
