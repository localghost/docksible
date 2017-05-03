package ansible

import (
	"github.com/localghost/docksible/docker"
	"strings"
)

type dockerConnector struct {
	containerId   string
	containerName string
}

func NewDockerConnector() Connector {
	return &dockerConnector{}
}

func (c *dockerConnector) Connect(source *docker.Container, target *docker.Container) error {
	c.containerId = target.Id
	// Container name starts with slash '/'. Maybe name should be taken directly from --hostname/-n instead of going
	// through inspect.
	c.containerName = strings.TrimLeft(target.Inspect().Name, "/")
	return nil
}

func (c *dockerConnector) Name() string {
	return "docker"
}

func (c *dockerConnector) Host() string {
	if c.containerName != "" {
		return c.containerName
	}
	return c.containerId
}

func (c *dockerConnector) ExtraArgs() []string {
	return []string{}
}

func (c *dockerConnector) Disconnect() error {
	return nil
}
