package provisioner

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"

	"github.com/localghost/docksible/utils"
	"log"
	"net/url"
	"strings"
)

type Provisioner struct {
	image string

	cli *client.Client
	ctx context.Context
}

func NewProvisioner(image string, cli *client.Client) *Provisioner {
	return &Provisioner{image: image, cli: cli, ctx: context.Background()}
}

func (b *Provisioner) Run(ansibleDir, playbookPath string) (*docker.Container, error) {
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

	// TODO: do it only in case DOCKER_HOST is not set
	dockerAddress, err := url.Parse(client.DefaultDockerHost)
	if err != nil {
		log.Fatal(err)
	}
	if utils.Exists(dockerAddress.Path) {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: dockerAddress.Path, Target: dockerAddress.Path},
		)
	}

	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      b.image,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{Mounts: mounts}
	return docker.NewContainer("", config, hostConfig, nil, b.cli)
}
