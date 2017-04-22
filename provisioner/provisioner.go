package provisioner

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"

	"log"
	"strings"
)

type Provisioner struct {
	image string

	cli *client.Client
	ctx context.Context
}

func NewProvisioner(image string) *Provisioner {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &Provisioner{image: image, cli: cli, ctx: context.Background()}
}

func (b *Provisioner) Run(ansibleDir, playbookPath string) *docker.Container {
	mounts := []mount.Mount{}
	if ansibleDir != "" {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: ansibleDir, Target: ansibleDir},
		)
	}
	if ansibleDir == "" || !strings.HasPrefix(playbookPath, ansibleDir) {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: playbookPath, Target: playbookPath},
		)
	}
	mounts = append(
		mounts,
		mount.Mount{Type: "bind", Source: "/var/run/docker.sock", Target: "/var/run/docker.sock"},
	)

	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      b.image,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{Mounts: mounts}
	return docker.NewContainer("", config, hostConfig, nil, b.cli)
}
