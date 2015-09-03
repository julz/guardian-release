package rundmc_test

import (
	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/rundmc"
	"github.com/cloudfoundry-incubator/guardian/rundmc/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Creating a container", func() {
	var (
		fakeRepo *fakes.FakeRepo

		czr *rundmc.Containerizer
	)

	BeforeEach(func() {
		fakeRepo = new(fakes.FakeRepo)
		czr = &rundmc.Containerizer{
			Repo: fakeRepo,
		}
	})

	It("delegates creation to the repo", func() {
		spec := gardener.DesiredContainerSpec{Handle: "foo"}
		Expect(czr.Create(spec)).To(Succeed())

		Expect(fakeRepo.CreateCallCount()).To(Equal(1))
		Expect(fakeRepo.CreateArgsForCall(0)).To(Equal(spec))
	})

	Context("when the container exists", func() {
		It("delegates running to the repo", func() {
			anActualContainer := new(fakes.FakeActualContainer)
			fakeRepo.LookupReturns(anActualContainer, nil)

			stdout := gbytes.NewBuffer()
			czr.Run(
				"fishfinger",
				garden.ProcessSpec{Path: "foo"},
				garden.ProcessIO{Stdout: stdout},
			)

			Expect(fakeRepo.LookupArgsForCall(0)).To(Equal("fishfinger"))
			Expect(anActualContainer.RunCallCount()).To(Equal(1))

			spec, io := anActualContainer.RunArgsForCall(0)
			Expect(spec.Path).To(Equal("foo"))
			Expect(io.Stdout).To(Equal(stdout))
		})
	})

	Context("when the container does not exist", func() {
		PIt("returns an error", func() {
		})
	})

	// Context("running a process", func() {
	// 	Context("whether or not create has been run", func() {
	// 		It("delegates to a process tracker", func() {
	// 			pt.RunStub = func(id uint32, cmd *exec.Cmd, io garden.ProcessIO, tty *garden.TTYSpec, signaller process_tracker.Signaller) (garden.Process, error) {
	// 				cmd.Stdout = io.Stdout
	// 				cmd.Stderr = io.Stderr
	// 				cmd.Run()

	// 				return nil, nil
	// 			}

	// 			var err error
	// 			stdout = gbytes.NewBuffer()
	// 			returnedProcess, err = czr.Run(gardener.Handle("a-handle"), garden.ProcessSpec{
	// 				Path: "sh", Args: []string{"-c", "echo hello; echo stderr >2"},
	// 			}, garden.ProcessIO{
	// 				Stdout: stdout,
	// 			})

	// 			Expect(err).NotTo(HaveOccurred())
	// 			Expect(pt.RunCallCount()).To(Equal(1))
	// 			Expect(stdout).To(gbytes.Say("hello"))
	// 		})
	// 	})
	// })
})
