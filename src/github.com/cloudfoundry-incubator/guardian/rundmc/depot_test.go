package rundmc_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/rundmc"
	"github.com/cloudfoundry-incubator/guardian/rundmc/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/opencontainers/specs"
)

var _ = Describe("The Depot", func() {
	var (
		dir        string
		rootfsPath string

		actualContainerProvider *fakes.FakeActualContainerProvider
		depot                   *rundmc.Depot
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "depot")
		Expect(err).NotTo(HaveOccurred())

		rootfsPath, err = ioutil.TempDir("", "rootfs")
		Expect(err).NotTo(HaveOccurred())

		actualContainerProvider = new(fakes.FakeActualContainerProvider)
		depot = &rundmc.Depot{
			Dir: dir,
			ActualContainerProvider: actualContainerProvider,
		}
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	Describe("Lookup", func() {
		Context("when the container directory exists", func() {
			It("returns an actual container from the directory", func() {
				Expect(os.Mkdir(path.Join(dir, "12345"), 0700)).To(Succeed())

				anActualContainer := new(fakes.FakeActualContainer)
				actualContainerProvider.ProvideReturns(anActualContainer, nil)

				container, err := depot.Lookup("12345")
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(anActualContainer))
				Expect(actualContainerProvider.ProvideArgsForCall(0)).To(Equal(path.Join(dir, "12345")))
			})
		})

		PContext("when the container directory does not exist", func() {
		})
	})

	Describe("Create", func() {
		BeforeEach(func() {
			spec := gardener.DesiredContainerSpec{
				Handle:     "hello",
				RootFSPath: rootfsPath,
			}

			Expect(depot.Create(spec)).To(Succeed())
		})

		Describe("a subdirectory of depot named after the handle of the container", func() {
			var theCreatedDirectory string

			BeforeEach(func() {
				theCreatedDirectory = path.Join(dir, "hello")
			})

			It("should be created", func() {
				Expect(theCreatedDirectory).To(BeADirectory())
			})

			Describe("the rootfs subdirectory", func() {
				It("should be symlinked to the rootfs directory from the spec", func() {
					rootfsSubdir := path.Join(theCreatedDirectory, "rootfs")
					Expect(rootfsSubdir).To(BeAnExistingFile())

					Expect(os.Readlink(rootfsSubdir)).To(Equal(rootfsPath))
				})
			})

			Describe("a config.json", func() {
				It("should be created", func() {
					Expect(path.Join(theCreatedDirectory, "config.json")).To(BeAnExistingFile())
				})

				Describe("the parsed json", func() {
					var spec *specs.Spec

					BeforeEach(func() {
						spec = &specs.Spec{}
						config, err := os.Open(path.Join(theCreatedDirectory, "config.json"))
						Expect(err).NotTo(HaveOccurred())
						json.NewDecoder(config).Decode(&spec)
					})

					It("should contain the current spec version", func() {
						Expect(spec.Version).To(Equal(specs.Version))
					})

					It("should set the roots to the rootfs subdirectory", func() {
						Expect(spec.Root.Path).To(Equal("rootfs"))
					})

					It("should set the initial process to a dummy process", func() {
						Expect(spec.Process).To(Equal(specs.Process{
							Terminal: true,
							Args:     []string{"sh"},
						}))
					})
				})
			})
		})
	})
})
