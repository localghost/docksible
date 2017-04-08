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

func (p *product) Run() string {
	p.buildProductImage()
	containerId := p.runProductContainer()
	p.provisionContainer(containerId)
	fmt.Println(containerId)
	return containerId
	//err := p.cli.ContainerStop(p.ctx, containerId, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
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

func (p *product) runProductContainer() string {
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
	response, err := p.cli.ContainerCreate(p.ctx, config, hostConfig, nil, "")
	if err != nil {
		log.Fatal(err)
	}
	err = p.cli.ContainerStart(p.ctx, response.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return response.ID
}

func (p *product) provisionContainer(containerId string) {

}
