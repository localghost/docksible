package ansible

import "github.com/localghost/docksible/docker"

type dockerConnector struct {
	containerId string
}

func NewDockerConnector() Connector {
	return &dockerConnector{}
}

func (c *dockerConnector) Connect(source *docker.Container, target *docker.Container) error {
	c.containerId = target.Id
	return nil
}

func (c *dockerConnector) Name() string {
	return "docker"
}

func (c *dockerConnector) Host() string {
	return c.containerId
}

func (c *dockerConnector) ExtraArgs() []string {
	return []string{}
}

func (c *dockerConnector) Disconnect() error {
	return nil
}
