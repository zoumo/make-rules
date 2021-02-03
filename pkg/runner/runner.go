package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Status struct {
	cmd    string
	output string
	err    error
}

func NewRunnerError(cmd, output string, err error) error {
	return &Status{
		cmd:    cmd,
		output: output,
		err:    err,
	}
}

func (e *Status) Error() string {
	return fmt.Sprintf("failed to run cmd: cmd=%s, err=%v, output=%s", e.cmd, e.err, e.output)
}

type Runner struct {
	name string
	env  map[string]string
	dir  string
}

func NewRunner(name string) *Runner {
	c := &Runner{
		name: name,
	}
	c.init()
	return c
}

func (c *Runner) init() {
	envs := os.Environ()
	if c.env == nil {
		c.env = make(map[string]string)
	}
	for _, env := range envs {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) != 2 {
			continue
		}
		c.env[kv[0]] = kv[1]
	}
	// overwrite env
	c.env["GO111MODULE"] = "on"
}

func (c *Runner) clone() *Runner {
	cc := &Runner{
		name: c.name,
		env:  map[string]string{},
		dir:  c.dir,
	}
	for k, v := range c.env {
		cc.env[k] = v
	}

	return cc
}

func (c *Runner) FilterEnv(requires []string) map[string]string {
	required := map[string]string{}
	for _, k := range requires {
		if v, ok := c.env[k]; ok {
			required[k] = v
		}
	}
	return required
}

func (c *Runner) WithDir(dir string) *Runner {
	cc := c.clone()
	cc.dir = dir
	return cc
}

func (c *Runner) WithEnvs(kvs ...string) *Runner {
	cc := c.clone()
	length := len(kvs)
	for i := 0; i < length; {
		if i+1 >= length {
			// skip this key
			break
		}
		k := kvs[i]
		v := kvs[i+1]
		cc.env[k] = v
		i += 2
	}
	return cc
}

func (c *Runner) cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(c.name, args...)
	// merge env
	cmd.Env = joinMap(c.env, "=")
	cmd.Dir = c.dir
	return cmd
}

func (c *Runner) RunOutput(args ...string) ([]byte, error) {
	cmd := c.cmd(args...)
	out, err := cmd.Output()
	if err != nil {
		output := out
		if eerr, ok := err.(*exec.ExitError); ok {
			output = eerr.Stderr
		}
		return nil, NewRunnerError(cmd.String(), string(output), err)
	}
	return out, err
}

func (c *Runner) RunCombinedOutput(args ...string) ([]byte, error) {
	cmd := c.cmd(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := out
		if eerr, ok := err.(*exec.ExitError); ok {
			output = append(output, eerr.Stderr...)
		}
		return nil, NewRunnerError(cmd.String(), string(output), err)
	}
	return out, err
}

func joinMap(m map[string]string, delimiter string) []string {
	out := make([]string, len(m))

	i := 0
	for k, v := range m {
		out[i] = k + delimiter + v
		i++
	}
	return out
}
