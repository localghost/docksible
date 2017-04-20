package builder

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"

	"fmt"
	"github.com/localghost/docksible/ansible"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type builder struct {
	image string

	cli       *client.Client
	ctx       context.Context
	container *docker.Container
	result    *docker.Container
}

type ProvisionOptions struct {
	PlaybookPath    string
	AnsibleDir      string
	InventoryGroups []string
}

func New(image string) *builder {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &builder{image: image, cli: cli, ctx: context.Background()}
}

func (b *builder) ProvisionContainer(container *docker.Container, options *ProvisionOptions) {
	mounts := []mount.Mount{}
	if options.AnsibleDir != "" {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: options.AnsibleDir, Target: options.AnsibleDir},
		)
	}
	if !filepath.IsAbs(options.PlaybookPath) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		options.PlaybookPath = filepath.Join(cwd, options.PlaybookPath)
	}
	if options.AnsibleDir == "" || !strings.HasPrefix(options.PlaybookPath, options.AnsibleDir) {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: options.PlaybookPath, Target: options.PlaybookPath},
		)
	}
	mounts = append(
		mounts,
		mount.Mount{Type: "bind", Source: "/var/run/docker.sock", Target: "/var/run/docker.sock"},
	)
	b.container = b.runBuilderContainer(mounts)
	defer b.container.StopAndRemove()
	b.result = container
	b.setupProvisionedContainer(options.PlaybookPath)
}

func (b *builder) setupProvisionedContainer(playbookPath string) error {
	ans := ansible.New(
		b.container,
		ansible.ExecutorFunc(func(command []string) error {
			code, err := b.container.ExecAndOutput(os.Stdout, os.Stderr, command...)
			if err != nil {
				log.Fatal(err)
			}
			if code != 0 {
				return fmt.Errorf("Command failed with %d", code)
			}
			return nil
		}),
		"",
	)
	err := ans.Play(playbookPath, ansible.PlayTarget{Container: b.result, Connector: ansible.NewDockerConnector()})
	//err := ans.Play(playbookPath, ansible.PlayTarget{Container: b.result, Connector: ansible.NewSshConnector()})
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (b *builder) runBuilderContainer(mounts []mount.Mount) *docker.Container {
	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      b.image,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{Mounts: mounts}
	return docker.NewContainer("", config, hostConfig, nil, nil)
}
