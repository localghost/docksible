package docker

import (
	"context"
	"io"
	"log"
	"strings"

	"io/ioutil"

	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Container struct {
	Id string

	ctx context.Context
	cli *client.Client
}

func NewContainer(name string, config *container.Config, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig, cli *client.Client) *Container {
	if cli == nil {
		var err error
		cli, err = client.NewEnvClient()
		if err != nil {
			log.Fatal(err)
		}
	}

	result := &Container{ctx: context.Background(), cli: cli}
	result.run(name, config, hostConfig, netConfig)
	return result
}

func (c *Container) run(name string, config *container.Config, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig) {
	c.pullImageIfNotExists(config.Image)

	response, err := c.cli.ContainerCreate(c.ctx, config, hostConfig, nil, "")
	if err != nil {
		log.Fatal(err)
	}
	c.Id = response.ID

	err = c.cli.ContainerStart(c.ctx, c.Id, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Container) pullImageIfNotExists(image string) {
	if !c.imageExists(image) {
		response, err := c.cli.ImagePull(c.ctx, image, types.ImagePullOptions{})
		if err != nil {
			log.Fatal(err)
		}
		defer response.Close()
		io.Copy(ioutil.Discard, response)
	}
}

func (c *Container) imageExists(image string) bool {
	filter := filters.NewArgs()
	filter.Add("reference", image)

	results, err := c.cli.ImageList(c.ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		log.Fatal(err)
	}

	return len(results) != 0
}

func (c *Container) Exec(command ...string) {
	log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
	response, err := c.cli.ContainerExecCreate(c.ctx, c.Id, types.ExecConfig{
		AttachStdout: false,
		Cmd:          command,
		Detach:       true,
		AttachStderr: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = c.cli.ContainerExecStart(c.ctx, response.ID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Container) ExecAndWait(command ...string) (int, error) {
	log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
	response, err := c.cli.ContainerExecCreate(c.ctx, c.Id, types.ExecConfig{
		AttachStdout: true,
		Cmd:          command,
		Detach:       false,
		AttachStderr: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	hijacked, err := c.cli.ContainerExecAttach(c.ctx, response.ID, types.ExecConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer hijacked.Close()
	io.Copy(ioutil.Discard, hijacked.Reader)

	inspect, err := c.cli.ContainerExecInspect(c.ctx, response.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return inspect.ExitCode, nil
}

func (c *Container) ExecAndOutput(stdout, stderr io.Writer, command ...string) (int, error) {
	log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
	response, err := c.cli.ContainerExecCreate(c.ctx, c.Id, types.ExecConfig{
		AttachStdout: true,
		Cmd:          command,
		Detach:       false,
		AttachStderr: true,
		Tty:          false, // enable to turn on coloring
	})
	if err != nil {
		log.Fatal(err)
	}
	hijacked, err := c.cli.ContainerExecAttach(c.ctx, response.ID, types.ExecConfig{})
	if err != nil {
		log.Fatal(err)
	}
	defer hijacked.Close()
	stdcopy.StdCopy(stdout, stderr, hijacked.Reader)

	inspect, err := c.cli.ContainerExecInspect(c.ctx, response.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return inspect.ExitCode, nil
}

func (c *Container) Commit(image string, command string) string {
	changes := []string{fmt.Sprintf("CMD %s", command)}
	response, err := c.cli.ContainerCommit(c.ctx, c.Id, types.ContainerCommitOptions{Reference: image, Changes: changes})
	if err != nil {
		log.Fatal(err)
	}
	return response.ID
}

func (c *Container) StopAndRemove() {
	if err := c.cli.ContainerStop(c.ctx, c.Id, nil); err != nil {
		log.Println(err)
	}
	c.Remove()
}

func (c *Container) Remove() {
	if err := c.cli.ContainerRemove(c.ctx, c.Id, types.ContainerRemoveOptions{RemoveVolumes: true}); err != nil {
		log.Fatal(err)
	}
}

func (c *Container) Inspect() types.ContainerJSON {
	result, err := c.cli.ContainerInspect(c.ctx, c.Id)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func (c *Container) CopyTo(dest string, content io.Reader) {
	err := c.cli.CopyToContainer(c.ctx, c.Id, dest, content, types.CopyToContainerOptions{})
	if err != nil {
		log.Fatal(err)
	}
}
