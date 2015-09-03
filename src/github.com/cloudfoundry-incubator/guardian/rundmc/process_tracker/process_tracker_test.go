package process_tracker_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/rundmc/process_tracker"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
)

var _ = Describe("Process tracker", func() {
	var (
		processTracker process_tracker.ProcessTracker
		tmpdir         string
		signaller      process_tracker.Signaller
	)

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "process-tracker-tests")
		Expect(err).ToNot(HaveOccurred())

		err = os.MkdirAll(filepath.Join(tmpdir, "bin"), 0755)
		Expect(err).ToNot(HaveOccurred())

		err = copyFile(iodaemonBin, filepath.Join(tmpdir, "bin", "iodaemon"))
		Expect(err).ToNot(HaveOccurred())

		signaller = &process_tracker.LinkSignaller{}

		iodaemon, err := gexec.Build("github.com/cloudfoundry-incubator/garden-linux/iodaemon/cmd/iodaemon")
		Expect(err).ToNot(HaveOccurred())

		processTracker = process_tracker.New(tmpdir, iodaemon, linux_command_runner.New())
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Describe("Running processes", func() {
		It("runs the process and returns its exit code", func() {
			cmd := exec.Command("bash", "-c", "exit 42")

			process, err := processTracker.Run(555, cmd, garden.ProcessIO{}, nil, signaller)
			Expect(err).NotTo(HaveOccurred())

			status, err := process.Wait()
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(42))
		})

		Describe("signalling a running process", func() {
			var (
				process garden.Process
				stdout  *gbytes.Buffer
				cmd     *exec.Cmd
			)

			JustBeforeEach(func() {
				var err error
				cmd = exec.Command(testPrintBin)

				stdout = gbytes.NewBuffer()
				process, err = processTracker.Run(
					2, cmd,
					garden.ProcessIO{
						Stdout: io.MultiWriter(stdout, GinkgoWriter),
						Stderr: GinkgoWriter,
					}, nil, signaller)
				Expect(err).NotTo(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("pid"))
			})

			AfterEach(func() {
				if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
					cmd.Process.Signal(os.Kill)
				}
			})

			Context("when the signaller is an LinkSignaller", func() {
				It("sends a kill message on the extra file descriptor", func(done Done) {
					Expect(process.Signal(garden.SignalKill)).To(Succeed())
					Eventually(stdout).Should(gbytes.Say("Received: killed"))
					close(done)
				}, 2.0)

				It("kills the process with a terminate signal", func(done Done) {
					Expect(process.Signal(garden.SignalTerminate)).To(Succeed())
					Eventually(stdout).Should(gbytes.Say("Received: terminated"))
					close(done)
				}, 2.0)

				Context("when an unsupported signal is sent", func() {
					AfterEach(func() {
						Expect(process.Signal(garden.SignalKill)).To(Succeed())
					})

					It("return error", func() {
						Expect(process.Signal(garden.Signal(999))).To(MatchError(HaveSuffix("failed to send signal: unknown signal: 999")))
					})
				})
			})
		})

		It("streams the process's stdout and stderr", func() {
			cmd := exec.Command(
				"/bin/bash",
				"-c",
				"echo 'hi out' && echo 'hi err' >&2",
			)

			stdout := gbytes.NewBuffer()
			stderr := gbytes.NewBuffer()

			_, err := processTracker.Run(40, cmd, garden.ProcessIO{
				Stdout: stdout,
				Stderr: stderr,
			}, nil, signaller)
			Expect(err).NotTo(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("hi out\n"))
			Eventually(stderr).Should(gbytes.Say("hi err\n"))
		})

		It("streams input to the process", func() {
			stdout := gbytes.NewBuffer()

			_, err := processTracker.Run(50, exec.Command("cat"), garden.ProcessIO{
				Stdin:  bytes.NewBufferString("stdin-line1\nstdin-line2\n"),
				Stdout: stdout,
			}, nil, signaller)
			Expect(err).NotTo(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("stdin-line1\nstdin-line2\n"))
		})

		Context("when there is an error reading the stdin stream", func() {
			It("does not close the process's stdin", func() {
				pipeR, pipeW := io.Pipe()
				stdout := gbytes.NewBuffer()

				process, err := processTracker.Run(60, exec.Command("cat"), garden.ProcessIO{
					Stdin:  pipeR,
					Stdout: stdout,
				}, nil, signaller)
				Expect(err).NotTo(HaveOccurred())

				pipeW.Write([]byte("Hello stdin!"))
				Eventually(stdout).Should(gbytes.Say("Hello stdin!"))

				pipeW.CloseWithError(errors.New("Failed"))
				Consistently(stdout, 0.1).ShouldNot(gbytes.Say("."))

				pipeR, pipeW = io.Pipe()
				processTracker.Attach(process.ID(), garden.ProcessIO{
					Stdin: pipeR,
				})

				pipeW.Write([]byte("Hello again, stdin!"))
				Eventually(stdout).Should(gbytes.Say("Hello again, stdin!"))

				pipeW.Close()
				exitStatus, err := process.Wait()
				Expect(err).ToNot(HaveOccurred())
				Expect(exitStatus).To(Equal(0))
			})

			It("supports attaching more than once", func() {
				pipeR, pipeW := io.Pipe()
				stdout := gbytes.NewBuffer()

				process, err := processTracker.Run(70, exec.Command("cat"), garden.ProcessIO{
					Stdin:  pipeR,
					Stdout: stdout,
				}, nil, signaller)
				Expect(err).NotTo(HaveOccurred())

				pipeW.Write([]byte("Hello stdin!"))
				Eventually(stdout).Should(gbytes.Say("Hello stdin!"))

				pipeW.CloseWithError(errors.New("Failed"))
				Consistently(stdout, 0.1).ShouldNot(gbytes.Say("."))

				pipeR, pipeW = io.Pipe()
				_, err = processTracker.Attach(process.ID(), garden.ProcessIO{
					Stdin: pipeR,
				})
				Expect(err).ToNot(HaveOccurred())

				pipeW.Write([]byte("Hello again, stdin!"))
				Eventually(stdout).Should(gbytes.Say("Hello again, stdin!"))

				pipeR, pipeW = io.Pipe()

				_, err = processTracker.Attach(process.ID(), garden.ProcessIO{
					Stdin: pipeR,
				})
				Expect(err).ToNot(HaveOccurred())

				pipeW.Write([]byte("Hello again again, stdin!"))
				Eventually(stdout, "1s").Should(gbytes.Say("Hello again again, stdin!"))

				pipeW.Close()
				Expect(process.Wait()).To(Equal(0))
			})
		})

		Context("with a tty", func() {
			It("forwards TTY signals to the process", func() {
				cmd := exec.Command("/bin/bash", "-c", `
				trap "stty size; exit 123" SIGWINCH
				stty size
				read
			`)

				stdout := gbytes.NewBuffer()

				process, err := processTracker.Run(90, cmd, garden.ProcessIO{
					Stdout: stdout,
				}, &garden.TTYSpec{
					WindowSize: &garden.WindowSize{
						Columns: 95,
						Rows:    13,
					},
				}, signaller)
				Expect(err).NotTo(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("13 95"))

				process.SetTTY(garden.TTYSpec{
					WindowSize: &garden.WindowSize{
						Columns: 101,
						Rows:    27,
					},
				})

				Eventually(stdout).Should(gbytes.Say("27 101"))
				Expect(process.Wait()).To(Equal(123))
			})

			Describe("when a window size is not specified", func() {
				It("picks a default window size", func() {
					cmd := exec.Command("/bin/bash", "-c", `
					stty size
				`)

					stdout := gbytes.NewBuffer()

					_, err := processTracker.Run(100, cmd, garden.ProcessIO{
						Stdout: stdout,
					}, &garden.TTYSpec{}, signaller)
					Expect(err).NotTo(HaveOccurred())

					Eventually(stdout).Should(gbytes.Say("24 80"))
				})
			})
		})

		Context("when spawning fails", func() {
			It("returns the error", func() {
				_, err := processTracker.Run(200, exec.Command("/bin/does-not-exist"), garden.ProcessIO{}, nil, signaller)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Restoring processes", func() {
		It("tracks the restored process", func() {
			processTracker.Restore(2, signaller)

			activeProcesses := processTracker.ActiveProcesses()
			Expect(activeProcesses).To(HaveLen(1))
			Expect(activeProcesses[0].ID()).To(Equal(uint32(2)))
		})
	})

	Describe("Attaching to running processes", func() {
		It("streams stdout, stdin, and stderr", func() {
			cmd := exec.Command("bash", "-c", `
			stuff=$(cat)
			echo "hi stdout" $stuff
			echo "hi stderr" $stuff >&2
		`)

			process, err := processTracker.Run(855, cmd, garden.ProcessIO{}, nil, signaller)
			Expect(err).NotTo(HaveOccurred())

			stdout := gbytes.NewBuffer()
			stderr := gbytes.NewBuffer()

			process, err = processTracker.Attach(process.ID(), garden.ProcessIO{
				Stdin:  bytes.NewBufferString("this-is-stdin"),
				Stdout: stdout,
				Stderr: stderr,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("hi stdout this-is-stdin"))
			Eventually(stderr).Should(gbytes.Say("hi stderr this-is-stdin"))
		})
	})

	Describe("Listing active process IDs", func() {
		It("includes running process IDs", func() {
			stdin1, stdinWriter1 := io.Pipe()
			stdin2, stdinWriter2 := io.Pipe()

			Expect(processTracker.ActiveProcesses()).To(BeEmpty())

			process1, err := processTracker.Run(9955, exec.Command("cat"), garden.ProcessIO{
				Stdin: stdin1,
			}, nil, signaller)
			Expect(err).ToNot(HaveOccurred())

			Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process1))

			process2, err := processTracker.Run(9956, exec.Command("cat"), garden.ProcessIO{
				Stdin: stdin2,
			}, nil, signaller)
			Expect(err).ToNot(HaveOccurred())

			Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process1, process2))

			stdinWriter1.Close()
			Eventually(processTracker.ActiveProcesses).Should(ConsistOf(process2))

			stdinWriter2.Close()
			Eventually(processTracker.ActiveProcesses).Should(BeEmpty())
		})
	})
})

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}

	defer s.Close()

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(d, s)
	if err != nil {
		d.Close()
		return err
	}

	return d.Close()
}
