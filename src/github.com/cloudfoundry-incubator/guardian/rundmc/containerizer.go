package rundmc

import (
	"os/exec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/rundmc/process_tracker"
)

//go:generate counterfeiter . ProcessTracker
type ProcessTracker interface {
	Run(id uint32, cmd *exec.Cmd, io garden.ProcessIO, tty *garden.TTYSpec, signaller process_tracker.Signaller) (garden.Process, error)
}

//go:generate counterfeiter . Repo
type Repo interface {
	Create(gardener.DesiredContainerSpec) error
	Lookup(string) (ActualContainer, error)
	Destroy(string) error
}

//go:generate counterfeiter . ActualContainer
type ActualContainer interface {
	Run(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error)
}

type Containerizer struct {
	Repo Repo
}

func (c *Containerizer) Create(spec gardener.DesiredContainerSpec) error {
	return c.Repo.Create(spec)
}

func (c *Containerizer) Run(handle string, spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	ac, _ := c.Repo.Lookup(handle)
	return ac.Run(spec, io)
}
