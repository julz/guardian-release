package rundmc_test

import (
	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/rundmc"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("RunDMC Integration", func() {
	XIt("runs a process in a container", func() {
		containerizer := rundmc.Containerizer{}

		containerizer.Create(gardener.DesiredContainerSpec{
			Handle: "the-handle",
		})
	})
})
