package gardener_test

import (
	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/gardener/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Gardener", func() {
	Context("with dummy implementations of the major components", func() {
		var (
			containerizer *fakes.FakeContainerizer
			volumizer     *fakes.FakeVolumizer
			networker     *fakes.FakeNetworker

			gdnr *gardener.Gardener
		)

		BeforeEach(func() {
			containerizer = new(fakes.FakeContainerizer)
			volumizer = new(fakes.FakeVolumizer)
			networker = new(fakes.FakeNetworker)

			volumizer.VolumizeReturns("the-volumized-rootfs-path", nil)

			containerizer.CreateStub = func(spec gardener.DesiredContainerSpec) error {
				return nil
			}

			gdnr = &gardener.Gardener{
				Networker:     networker,
				Volumizer:     volumizer,
				Containerizer: containerizer,
			}
		})

		Describe("creating a container", func() {
			var (
				createdContainer garden.Container
				spec             garden.ContainerSpec
			)

			BeforeEach(func() {
				spec = garden.ContainerSpec{}
			})

			JustBeforeEach(func() {
				var err error
				createdContainer, err = gdnr.Create(spec)

				Expect(err).NotTo(HaveOccurred())
			})

			It("passes the rootfs provided by the volumizer to the containerizer", func() {
				Expect(containerizer.CreateArgsForCall(0).RootFSPath).To(Equal("the-volumized-rootfs-path"))
			})

			PIt("destroys volumes when creating fails", func() {})

			Context("when a handle is specified", func() {
				BeforeEach(func() {
					spec.Handle = "the-handle"
				})

				It("passes the handle to the containerizer", func() {
					Expect(containerizer.CreateArgsForCall(0).Handle).To(Equal("the-handle"))
				})

				It("remembers the handle", func() {
					Expect(createdContainer.Handle()).To(Equal("the-handle"))
				})

				PContext("and it is already in use", func() {})
			})

			Context("when a handle is not specified", func() {
				PIt("assigns a handle", func() {
				})
			})

			PContext("when creating the container fails", func() {
			})

			Context("looking up a container and running a process", func() {
				var (
					stdout *gbytes.Buffer
					stderr *gbytes.Buffer
				)

				BeforeEach(func() {
					stdout = gbytes.NewBuffer()
					stderr = gbytes.NewBuffer()

					containerizer.RunStub = func(container string, spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
						io.Stdout.Write([]byte("stdout"))
						io.Stderr.Write([]byte("stderr"))
						return nil, nil
					}
				})

				JustBeforeEach(func() {
					lookedupContainer, err := gdnr.Lookup(createdContainer.Handle())
					Expect(err).NotTo(HaveOccurred())

					lookedupContainer.Run(garden.ProcessSpec{
						Path: "foo",
						Args: []string{"bar", "baz"},
						Env:  []string{"baked", "beans"},
					}, garden.ProcessIO{
						Stdout: stdout,
						Stderr: stderr,
					})
				})

				It("streams stdout", func() {
					Expect(stdout).To(gbytes.Say("stdout"))
				})

				It("streams stdout", func() {
					Expect(stderr).To(gbytes.Say("stderr"))
				})
			})
		})
	})
})
