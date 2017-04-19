package builder

import (
	//"encoding/json"
	//"io"
	//"bufio"
	//"bytes"
	"context"
	//"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/utils"
	//"io"

	"fmt"
	"github.com/localghost/docksible/ansible"
	"log"
	"os"
	"path/filepath"
)

type builder struct {
	image string

	cli       *client.Client
	ctx       context.Context
	container *docker.Container
	result    *docker.Container
}

type ProvisionOptions struct {
	PlaybookPath    string
	AnsibleDir      string
	InventoryGroups []string
}

func New(image string) *builder {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	return &builder{image: image, cli: cli, ctx: context.Background()}
}

//func (b *builder) Bootstrap() {
//	b.buildBuilderImage()
//}

func (b *builder) ProvisionContainer(container *docker.Container, options *ProvisionOptions) {
	mounts := []mount.Mount{}
	if options.AnsibleDir != "" {
		mounts = append(
			mounts,
			mount.Mount{Type: "bind", Source: options.AnsibleDir, Target: options.AnsibleDir},
		)
	}
	if !filepath.IsAbs(options.PlaybookPath) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		options.PlaybookPath = filepath.Join(cwd, options.PlaybookPath)
	}
	mounts = append(
		mounts,
		mount.Mount{Type: "bind", Source: options.PlaybookPath, Target: options.PlaybookPath},
	)
	mounts = append(
		mounts,
		mount.Mount{Type: "bind", Source: "/var/run/docker.sock", Target: "/var/run/docker.sock"},
	)
	b.container = b.runBuilderContainer(mounts)
	defer b.container.StopAndRemove()
	b.result = container
	b.setupProvisionedContainer(options.PlaybookPath)
}

//func (b *builder) buildBuilderImage() {
//	fmt.Println("Building builder image")
//	buildContext := b.createBuildContext()
//	buildOptions := types.ImageBuildOptions{
//		Tags: []string{"docksible-builder"},
//	}
//	response, err := b.cli.ImageBuild(b.ctx, buildContext, buildOptions)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer response.Body.Close()
//
//	type Line struct{ Stream string }
//
//	decoder := json.NewDecoder(response.Body)
//	for err == nil {
//		var line Line
//		if err = decoder.Decode(&line); err != nil {
//			break
//		}
//		fmt.Print(line.Stream)
//	}
//	io.Copy(os.Stdout, response.Body)
//}

//func (b *builder) createBuildContext() *bytes.Buffer {
//	buildContext := utils.NewInMemoryArchive()
//	buildContext.Add("Dockerfile", dockerfile)
//	return buildContext.Close()
//}

func (b *builder) setupProvisionedContainer(playbookPath string) error {
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
	_ = containerAddress

	//ssh := connector.NewSsh(containerAddress, "root", "/tmp/id_rsa")
	//ssh.Execute(
	//	connector.ExecutorFunc(func(command []string) error {
	//		code, err := b.container.ExecAndOutput(os.Stdout, os.Stderr, command...)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		return fmt.Errorf("Command failed eith %d", code)
	//	}),
	//	playbookPath,
	//)

	docker := ansible.NewDockerConnector(b.result.Id, "root")
	docker.Execute(
		ansible.ExecutorFunc(func(command []string) error {
			code, err := b.container.ExecAndOutput(os.Stdout, os.Stderr, command...)
			if err != nil {
				log.Fatal(err)
			}
			return fmt.Errorf("Command failed eith %d", code)
		}),
		playbookPath,
	)

	//resultCode, err := b.container.ExecAndOutput(os.Stdout, os.Stderr, cmd...)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (b *builder) runBuilderContainer(mounts []mount.Mount) *docker.Container {
	config := &container.Config{
		Cmd: []string{
			"bash", "-c", "tail -f /dev/null",
		},
		Image:      b.image,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{Mounts: mounts}
	return docker.NewContainer("", config, hostConfig, nil, nil)
}
