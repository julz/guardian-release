package rundmc

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/opencontainers/specs"
)

type RuncContainerFactory struct {
	Tracker ProcessTracker
}

func (a *RuncContainerFactory) Provide(containerDir string) (ActualContainer, error) {
	return &runcc{
		Dir:     containerDir,
		Tracker: a.Tracker,
	}, nil
}

type runcc struct {
	Dir     string
	Tracker ProcessTracker
}

func (c *runcc) Run(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	c.create()
	return c.exec(spec, io)
}

func (c *runcc) create() error {
	cmd := exec.Command("runc")
	cmd.Dir = c.Dir

	_, err := c.Tracker.Run(0, cmd, garden.ProcessIO{}, nil, nil)
	return err
}

func (c *runcc) exec(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	cmd, err := c.execCmd(spec)
	if err != nil {
		return nil, err
	}

	return c.Tracker.Run(0, cmd, io, spec.TTY, nil)
}

func (c *runcc) execCmd(spec garden.ProcessSpec) (*exec.Cmd, error) {
	jsonFile, err := ioutil.TempFile("", "rundmc-process-json")
	if err != nil {
		return nil, err
	}

	if err := json.NewEncoder(jsonFile).Encode(processSpec(spec)); err != nil {
		return nil, err
	}

	cmd := exec.Command("echo", "hello")
	cmd.Dir = c.Dir

	return cmd, nil
}

func processSpec(spec garden.ProcessSpec) specs.Process {
	return specs.Process{
		Args: append([]string{spec.Path}, spec.Args...),
		Env:  spec.Env,
		//User: spec.User, ??? wants a UID..
	}
}
