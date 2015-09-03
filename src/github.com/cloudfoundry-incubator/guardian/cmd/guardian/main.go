package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/garden-linux/logging"
	"github.com/cloudfoundry-incubator/garden/server"
	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/cloudfoundry-incubator/guardian/gardenshed"
	"github.com/cloudfoundry-incubator/guardian/rundmc"
	"github.com/cloudfoundry-incubator/guardian/rundmc/process_tracker"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager"
)

func main() {
	depotDir := flag.String(
		"depot",
		"/var/vcap/data/garden-runc/depot",
		"the depot directory to store containers in",
	)

	listenNetwork := flag.String(
		"listenNetwork",
		"tcp",
		"how to listen on the address (unix, tcp, etc.)",
	)

	listenAddr := flag.String(
		"listenAddr",
		"0.0.0.0:7777",
		"address to listen on",
	)

	containerGraceTime := flag.Duration(
		"containerGraceTime",
		0,
		"time after which to destroy idle containers",
	)

	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cf_lager.New("garden-runc")
	runner := &logging.Runner{
		CommandRunner: linux_command_runner.New(),
		Logger:        logger,
	}

	iodaemonBin, err := gexec.Build("github.com/cloudfoundry-incubator/garden-linux/iodaemon/cmd/iodaemon")
	if err != nil {
		panic(err)
	}

	gdnr := &gardener.Gardener{
		Volumizer: &gardenshed.Shed{},
		Containerizer: &rundmc.Containerizer{
			Repo: rundmc.Depot{
				Dir: *depotDir,
				ActualContainerProvider: &rundmc.RuncContainerFactory{
					Tracker: process_tracker.New("/tmp", iodaemonBin, runner),
				},
			},
		},
	}

	server := server.New(*listenNetwork, *listenAddr, *containerGraceTime, gdnr, logger)
	if err := server.Start(); err != nil {
		logger.Fatal("failed-to-start-server", err)
	}

	logger.Info("started", lager.Data{
		"network": *listenNetwork,
		"addr":    *listenAddr,
	})

	signals := make(chan os.Signal, 1)

	go func() {
		<-signals
		server.Stop()
		os.Exit(0)
	}()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	select {}
}
