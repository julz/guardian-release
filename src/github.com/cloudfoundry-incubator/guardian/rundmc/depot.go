package rundmc

import (
	"encoding/json"
	"os"
	"path"

	"github.com/cloudfoundry-incubator/guardian/gardener"
	"github.com/opencontainers/specs"
)

//go:generate counterfeiter . ActualContainerProvider
type ActualContainerProvider interface {
	Provide(directory string) (ActualContainer, error)
}

// a rundmc.Repo based on a directory. Each container is a subdirectory of the depot dir.
type Depot struct {
	ActualContainerProvider ActualContainerProvider
	Dir                     string
}

func (d Depot) Lookup(handle string) (ActualContainer, error) {
	return d.ActualContainerProvider.Provide(
		path.Join(d.Dir, handle),
	)
}

func (d Depot) Create(spec gardener.DesiredContainerSpec) error {
	depotDir := path.Join(d.Dir, spec.Handle)
	if err := createDepotDir(depotDir); err != nil {
		return err
	}

	if err := createConfigJSON(path.Join(depotDir, "config.json"), spec); err != nil {
		return err
	}

	if err := createRootfsSubdir(path.Join(depotDir, "rootfs"), spec.RootFSPath); err != nil {
		return err
	}

	return nil
}

func (d Depot) Destroy(handle string) error {
	return nil
}

func createDepotDir(dir string) error {
	return os.Mkdir(dir, 0700)
}

func createRootfsSubdir(subDir, rootfsPath string) error {
	return os.Symlink(rootfsPath, subDir)
}

func createConfigJSON(path string, spec gardener.DesiredContainerSpec) error {
	configJson, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(configJson).Encode(createSpec(spec)); err != nil {
		return err
	}

	return nil
}

func createSpec(spec gardener.DesiredContainerSpec) specs.Spec {
	return specs.Spec{
		Version: "pre-draft",
		Process: specs.Process{
			Terminal: true,
			Args:     []string{"sh"},
		},
		Root: specs.Root{
			Path: "rootfs",
		},
	}
}
