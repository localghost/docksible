package product

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"
	"io"
	"io/ioutil"
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
	p.buildProductImage()
	return p.runProductContainer()
}

func (p *product) buildProductImage() {
	filter := filters.NewArgs()
	filter.Add("reference", p.baseImage)

	results, err := p.cli.ImageList(p.ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		response, err := p.cli.ImagePull(p.ctx, p.baseImage, types.ImagePullOptions{})
		if err != nil {
			log.Fatal(err)
		}
		defer response.Close()
		io.Copy(ioutil.Discard, response)
	}
}

func (p *product) runProductContainer() *docker.Container {
	config := &container.Config{
		Cmd:   []string{"/usr/sbin/sshd", "-D"},
		Image: p.baseImage,
	}
	return docker.NewContainer("", config, nil, nil, nil)
}
