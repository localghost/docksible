package builder

import (
	"encoding/json"
	"io"
	//"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/utils"
	//"io"

	"log"
	"os"
)

const dockerfile string = `
FROM centos:7

RUN yum install -y epel-release
RUN yum install -y ansible
RUN yum install -y openssh-clients
`

type builder struct {
	builderImage string

	cli       *client.Client
	ctx       context.Context
	container *docker.Container
	result    *docker.Container
}

func New() *builder {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &builder{
		builderImage: "centos:7",
		cli:          cli,
		ctx:          context.Background(),
	}
}

func (b *builder) Bootstrap() {
	b.buildBuilderImage()
}

func (b *builder) ProvisionContainer(container *docker.Container) {
	b.container = b.runBuilderContainer()
	b.result = container
	b.setupProvisionedContainer()
}

func (b *builder) buildBuilderImage() {
	fmt.Println("Building builder image")
	buildContext := b.createBuildContext()
	buildOptions := types.ImageBuildOptions{
		Tags: []string{"docksible-builder"},
	}
	response, err := b.cli.ImageBuild(b.ctx, buildContext, buildOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	type Line struct{ Stream string }

	decoder := json.NewDecoder(response.Body)
	for err == nil {
		var line Line
		if err = decoder.Decode(&line); err != nil {
			break
		}
		fmt.Print(line.Stream)
	}
	io.Copy(os.Stdout, response.Body)
}

func (b *builder) createBuildContext() *bytes.Buffer {
	buildContext := utils.NewInMemoryArchive()
	buildContext.Add("Dockerfile", dockerfile)
	return buildContext.Close()
}

func (b *builder) setupProvisionedContainer() {
	sshKeys := utils.NewSSHKeyGenerator().GenerateInMemory()

	bc := utils.NewInMemoryArchive()
	bc.AddBytes("id_rsa", sshKeys.PrivateKey.Bytes())
	privateKey := bc.Close()
	b.cli.CopyToContainer(b.ctx, b.container.Id, "/tmp", privateKey, types.CopyToContainerOptions{})

	bc = utils.NewInMemoryArchive()
	bc.AddBytes("id_rsa.pub", sshKeys.PublicKey.Bytes())
	publicKey := bc.Close()
	err := b.cli.CopyToContainer(b.ctx, b.result.Id, "/tmp", publicKey, types.CopyToContainerOptions{})
	if err != nil {
		log.Fatal(err)
	}
	b.result.ExecAndWait("mkdir", "-p", "/root/.ssh")
	b.result.ExecAndWait("bash", "-c", "cat /tmp/id_rsa.pub >> /root/.ssh/authorized_keys")

	response, err := b.cli.ContainerInspect(b.ctx, b.result.Id)
	if err != nil {
		log.Fatal(err)
	}
	containerAddress := response.NetworkSettings.IPAddress
	cmd := []string{
		"/usr/bin/ansible-playbook",
		"/opt/ansible/playbook.yml",
		"-i", containerAddress + ",",
		"-l", containerAddress,
		"-vv",
		"--ssh-extra-args", "-o StrictHostKeyChecking=no -o IdentityFile=/tmp/id_rsa",
	}
	b.container.ExecAndOutput(os.Stdout, os.Stderr, cmd...)
}

func (b *builder) runBuilderContainer() *docker.Container {
	config := &container.Config{
		Cmd: []string{
			"bash", "-c", "tail -f /dev/null",
		},
		Image: "docksible-builder",
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
