package gqt

import (
	"fmt"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/gqt/runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
)

var _ = Describe("A simple container", func() {
	It("can run a process and get the output", func() {
		client := startGarden()
		container, err := client.Create(garden.ContainerSpec{
			Handle: "a-handle",
		})
		Expect(err).NotTo(HaveOccurred())

		stdout := gbytes.NewBuffer()
		container.Run(garden.ProcessSpec{
			Path: "echo",
			Args: []string{"hello"},
		}, garden.ProcessIO{Stdout: stdout, Stderr: GinkgoWriter})

		Eventually(stdout).Should(gbytes.Say("hello"))
	})
})

func startGarden(argv ...string) garden.Client {
	gardenBin, err := gexec.Build("github.com/cloudfoundry-incubator/guardian/cmd/guardian")
	Expect(err).NotTo(HaveOccurred())

	gardenAddr := fmt.Sprintf("/tmp/garden_%d.sock", GinkgoParallelNode())
	gardenRunner := runner.New("unix", gardenAddr, gardenBin, argv...)
	ifrit.Invoke(gardenRunner)

	return gardenRunner.NewClient()
}
