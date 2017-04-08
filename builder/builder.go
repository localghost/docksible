package builder

import (
	"archive/tar"
	//"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	//"io"
	"log"
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

	cli *client.Client
	ctx context.Context
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

func (b *builder) ProvisionContainer(containerId string) {
	builderContainerId := b.runBuilderContainer()
	b.setupProvisionedContainer(builderContainerId, containerId)
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
	response.Body.Close()
	fmt.Println("Image built")
}

func (b *builder) createBuildContext() *bytes.Buffer {
	buf := new(bytes.Buffer)

	writer := tar.NewWriter(buf)
	writer.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile))})
	writer.Write([]byte(dockerfile))
	writer.Close()

	return buf
}

func runCmd(command string, args ...string) {
	err := exec.Command(command, args...).Run()
	if err != nil {
		log.Fatal(err, command, args)
	}
}
func (b *builder) setupProvisionedContainer(builderContainerId, productContainerId string) {
	runCmd("/usr/bin/docker", "exec", builderContainerId, "mkdir", "-p", "/root/.ssh")
	runCmd("/usr/bin/docker", "exec", productContainerId, "mkdir", "-p", "/root/.ssh")
	runCmd("/usr/bin/docker", "exec", builderContainerId, "ssh-keygen", "-t", "rsa", "-N", "", "-q", "-f", "/root/.ssh/id_rsa")
	runCmd("/usr/bin/docker", "exec", builderContainerId, "bash", "-c", "chmod 0400 /root/.ssh/id_rsa*")
	runCmd("/usr/bin/docker", "cp", builderContainerId+":/root/.ssh/id_rsa.pub", "/tmp/authorized_keys")
	runCmd("/usr/bin/docker", "cp", "/tmp/authorized_keys", productContainerId+":/root/.ssh/authorized_keys")
	response, err := b.cli.ContainerInspect(b.ctx, productContainerId)
	if err != nil {
		log.Fatal(err)
	}
	containerAddress := response.NetworkSettings.IPAddress
	fmt.Println(containerAddress)
	runCmd("/usr/bin/docker", "exec", builderContainerId, "/usr/bin/ansible-playbook",
		"/opt/ansible/playbook.yml",
		"-i", containerAddress+",",
		"-l", containerAddress,
		"-vv",
		"--ssh-extra-args", "-o StrictHostKeyChecking=no",
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
