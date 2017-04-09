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
	//"io"

	"log"
	"os"
	"os/exec"
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
	builderContainerId := b.runBuilderContainer()
	b.container = docker.NewContainer(builderContainerId, nil)
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
	buildContext := docker.NewInMemoryBuildContext()
	buildContext.Add("Dockerfile", dockerfile)
	return buildContext.Close()
}

func runCmd(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err, command, args)
	}
}

func (b *builder) setupProvisionedContainer() {
	b.container.ExecAndWait("ssh-keygen", "-t", "rsa", "-N", "", "-q", "-f", "/tmp/id_rsa")
	b.container.ExecAndWait("bash", "-c", "chmod 0400 /tmp/id_rsa*")
	b.container.ExecAndWait("bash", "-c", "ls --color=never /opt")
	reader, _, err := b.cli.CopyFromContainer(b.ctx, b.container.Id, "/tmp/id_rsa.pub")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	fmt.Println("Setting SSH key as authorized")
	err = b.cli.CopyToContainer(b.ctx, b.result.Id, "/tmp", reader, types.CopyToContainerOptions{})
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
	b.container.ExecAndOutput(
		"/usr/bin/ansible-playbook",
		"/opt/ansible/playbook.yml",
		"-i", containerAddress+",",
		"-l", containerAddress,
		"-vv",
		"--ssh-extra-args", "-o StrictHostKeyChecking=no -o IdentityFile=/tmp/id_rsa",
	)
}

func (b *builder) runBuilderContainer() string {
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
	response, err := b.cli.ContainerCreate(b.ctx, config, hostConfig, nil, "")
	if err != nil {
		log.Fatal(err)
	}
	err = b.cli.ContainerStart(b.ctx, response.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return response.ID
}
