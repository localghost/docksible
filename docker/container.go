package docker

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Container struct {
	Id string

	ctx context.Context
	cli *client.Client
}

func NewContainer(id string, cli *client.Client) *Container {
	if cli == nil {
		var err error
		cli, err = client.NewEnvClient()
		if err != nil {
			log.Fatalln(err)
		}
	}
	return &Container{Id: id, ctx: context.Background(), cli: cli}
}

func (c *Container) ExecAndWait(command ...string) error {
	log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
	response, err := c.cli.ContainerExecCreate(c.ctx, c.Id, types.ExecConfig{
		AttachStdout: true,
		Cmd:          command,
		Detach:       false,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	hijacked, err := c.cli.ContainerExecAttach(c.ctx, response.ID, types.ExecConfig{})
	if err != nil {
		log.Fatalln(err)
	}
	defer hijacked.Close()
	io.Copy(ioutil.Discard, hijacked.Reader)

	inspect, err := c.cli.ContainerExecInspect(c.ctx, response.ID)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return nil
}

func (c *Container) ExecAndOutput(command ...string) error {
	log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
	response, err := c.cli.ContainerExecCreate(c.ctx, c.Id, types.ExecConfig{
		AttachStdout: true,
		Cmd:          command,
		Detach:       false,
		AttachStderr: true,
		// Tty:          true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	hijacked, err := c.cli.ContainerExecAttach(c.ctx, response.ID, types.ExecConfig{})
	if err != nil {
		log.Fatalln(err)
	}
	defer hijacked.Close()
	io.Copy(os.Stdout, hijacked.Reader)

	inspect, err := c.cli.ContainerExecInspect(c.ctx, response.ID)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return nil
}
