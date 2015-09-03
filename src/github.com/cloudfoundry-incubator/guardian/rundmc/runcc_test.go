package rundmc_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/rundmc"
	"github.com/cloudfoundry-incubator/guardian/rundmc/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/specs"
)

var _ = Describe("RuncContainerFactory", func() {
	var (
		containerDir string

		tracker *fakes.FakeProcessTracker
		factory rundmc.RuncContainerFactory
	)

	BeforeEach(func() {
		var err error
		containerDir, err = ioutil.TempDir("", "cdir")
		Expect(err).NotTo(HaveOccurred())

		tracker = new(fakes.FakeProcessTracker)
		factory = rundmc.RuncContainerFactory{
			Tracker: tracker,
		}
	})

	Context("after looking up a container", func() {
		var container rundmc.ActualContainer

		BeforeEach(func() {
			var err error
			container, err = factory.Provide(containerDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Running a Process", func() {
			var (
				theRunCmd *exec.Cmd
			)

			BeforeEach(func() {
				container.Run(garden.ProcessSpec{
					Path: "echo",
					Args: []string{"foo", "bar"},
					Env:  []string{"the", "environment", ",", "man"},
				}, garden.ProcessIO{})

			})

			Context("when the container does not exist yet", func() {
				It("creates the container", func() {
					Expect(tracker.RunCallCount()).To(BeNumerically(">", 0))
					_, theRunCmd, _, _, _ = tracker.RunArgsForCall(0)

					Expect(theRunCmd.Dir).To(Equal(containerDir))
					Expect(theRunCmd.Args).To(Equal([]string{"runc"}))
				})

				Context("after creating the container", func() {
					PIt("execs the process", func() {
					})
				})
			})

			Context("when the container already exists", func() {
				PIt("does not create it again", func() {
					Expect(tracker.RunCallCount()).To(Equal(1))
				})

				It("runs a process by executing `runC exec` in the container directory", func() {
					Expect(tracker.RunCallCount()).To(BeNumerically(">", 0))
					_, theRunCmd, _, _, _ = tracker.RunArgsForCall(1)

					Expect(theRunCmd.Dir).To(Equal(containerDir))
					Expect(theRunCmd.Args).To(HaveLen(3)) // runc exec [process.json]
					Expect(theRunCmd.Args[1]).To(Equal("exec"))
				})

				It("creates and passes a correct process spec as json to `runc exec`", func() {
					Expect(tracker.RunCallCount()).To(BeNumerically(">", 0))
					_, theRunCmd, _, _, _ = tracker.RunArgsForCall(1)

					Expect(theRunCmd.Args).To(HaveLen(3)) // runc exec [process.json]

					jsonFile, err := os.Open(theRunCmd.Args[2])
					Expect(err).NotTo(HaveOccurred())

					var spec specs.Process
					Expect(json.NewDecoder(jsonFile).Decode(&spec)).To(Succeed())

					Expect(spec.Args).To(Equal([]string{"echo", "foo", "bar"}))
					Expect(spec.Env).To(Equal([]string{"the", "environment", ",", "man"}))
				})

				PIt("cleans up the process spec json file", func() {})
			})
		})
	})
})
