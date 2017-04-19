package product

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"
	"log"
)

type product struct {
	baseImage string

	cli *client.Client
	ctx context.Context
}

func New(baseImage string) *product {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &product{baseImage: baseImage, cli: cli, ctx: context.Background()}
}

func (p *product) Run() *docker.Container {
	return p.runProductContainer()
}

func (p *product) runProductContainer() *docker.Container {
	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      p.baseImage,
		StopSignal: "SIGKILL",
	}
	return docker.NewContainer("", config, nil, nil, nil)
}
