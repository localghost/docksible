package product

import (
	"archive/tar"
	//"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	//"io"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/localghost/docksible/docker"
	"io"
	"io/ioutil"
	"log"
)

const dockerfile string = `
FROM centos:7

RUN yum install -y openssh-server
RUN /usr/sbin/sshd-keygen

`

type product struct {
	cli *client.Client
	ctx context.Context
}

func New() *product {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &product{
		cli: cli,
		ctx: context.Background(),
	}
}

func (p *product) Run() *docker.Container {
	p.buildProductImage()
	return p.runProductContainer()
}

func (p *product) buildProductImage() {
	fmt.Println("Building product image")
	buildContext := p.createBuildContext()
	buildOptions := types.ImageBuildOptions{
		Tags: []string{"docksible-product"},
	}
	response, err := p.cli.ImageBuild(p.ctx, buildContext, buildOptions)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(ioutil.Discard, response.Body)
	response.Body.Close()
	fmt.Println("Image built")
}

func (p *product) createBuildContext() *bytes.Buffer {
	buf := new(bytes.Buffer)

	writer := tar.NewWriter(buf)
	writer.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile))})
	writer.Write([]byte(dockerfile))
	writer.Close()

	return buf
}

func (p *product) runProductContainer() *docker.Container {
	config := &container.Config{
		Cmd:   []string{"/usr/sbin/sshd", "-D"},
		Image: "docksible-product",
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Source: "/tmp/ansible",
				Target: "/opt/ansible",
			},
		},
		AutoRemove: true,
	}
	return docker.NewContainer("", config, hostConfig, nil, nil)
}
