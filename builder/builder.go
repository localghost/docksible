package builder

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"

	"log"
	"os"
	"path/filepath"
	"strings"
)

type builder struct {
	image string

	cli *client.Client
	ctx context.Context
}

func New(image string) *builder {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &builder{image: image, cli: cli, ctx: context.Background()}
}

func (b *builder) Run(ansibleDir, playbookPath string) *docker.Container {
	mounts := []mount.Mount{}
	if ansibleDir != "" {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: ansibleDir, Target: ansibleDir},
		)
	}
	if !filepath.IsAbs(playbookPath) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		playbookPath = filepath.Join(cwd, playbookPath)
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
	return b.runBuilderContainer(mounts)
}

func (b *builder) runBuilderContainer(mounts []mount.Mount) *docker.Container {
	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      b.image,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{Mounts: mounts}
	return docker.NewContainer("", config, hostConfig, nil, b.cli)
}
