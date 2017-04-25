package docker

import (
	"context"
	"io"
	"log"

	"io/ioutil"

	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/localghost/docksible/utils"
	"path/filepath"
)

type Container struct {
	Id string

	ctx context.Context
	cli *client.Client
}

// TODO: Pass stream for communication with user.
func NewContainer(
	name string,
	config *container.Config,
	hostConfig *container.HostConfig,
	netConfig *network.NetworkingConfig,
	cli *client.Client,
) (*Container, error) {
	result := &Container{ctx: context.Background(), cli: cli}
	if err := result.run(name, config, hostConfig, netConfig); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Container) run(name string, config *container.Config, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig) error {
	if err := c.pullImageIfNotExists(config.Image); err != nil {
		return err
	}

	response, err := c.cli.ContainerCreate(c.ctx, config, hostConfig, nil, "")
	if err != nil {
		return err
	}
	c.Id = response.ID

	err = c.cli.ContainerStart(c.ctx, c.Id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) pullImageIfNotExists(image string) error {
	imageExists, err := c.imageExists(image)
	if err != nil {
		return err
	}

	if !imageExists {
		utils.Println("Image %s does not exist locally. Pulling it.", image)
		response, err := c.cli.ImagePull(c.ctx, image, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer response.Close()
		// TODO: Unpack the response and print it, check error.
		io.Copy(ioutil.Discard, response)
	}
	return nil
}

func (c *Container) imageExists(image string) (bool, error) {
	filter := filters.NewArgs()
	filter.Add("reference", image)

	results, err := c.cli.ImageList(c.ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		return false, err
	}

	return len(results) != 0, nil
}

func (c *Container) Exec(command ...string) {
	//log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
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
	//log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
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
	//log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return inspect.ExitCode, nil
}

func (c *Container) ExecAndOutput(stdout, stderr io.Writer, command ...string) (int, error) {
	//log.Printf("Running in %s: %s\n", c.Id, strings.Join(command, " "))
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
	//log.Printf("Running %v, ExitCode: %d\n", inspect.Running, inspect.ExitCode)
	return inspect.ExitCode, nil
}

func (c *Container) Commit(image string, command string) (string, error) {
	changes := []string{fmt.Sprintf("CMD %s", command)}
	response, err := c.cli.ContainerCommit(c.ctx, c.Id, types.ContainerCommitOptions{Reference: image, Changes: changes})
	if err != nil {
		return "", err
	}
	return response.ID, nil
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

// Copies content into the container, content has to be an archive.
func (c *Container) CopyTo(dest string, content io.Reader) {
	err := c.cli.CopyToContainer(c.ctx, c.Id, dest, content, types.CopyToContainerOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

// Copies content into the container, content is plain data.
func (c *Container) CopyContentTo(dest string, content io.Reader) {
	bc := utils.NewInMemoryArchive()
	bc.AddReader(filepath.Base(dest), content)
	buffer := bc.Close()

	c.CopyTo(filepath.Dir(dest), buffer)
}
