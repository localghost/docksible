package ansible

import "github.com/localghost/docksible/docker"

type dockerConnector struct {
	container string
	user      string
}

func NewDockerConnector(container string, user string) Connector {
	return &dockerConnector{container: container, user: user}
}

func (c *dockerConnector) Execute(executor Executor, playbook string) error {
	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-c", "docker",
		"-i", c.container + ",",
		"-l", c.container,
		"-e ansible_user=" + c.user,
		"-vv",
	}
	return executor.Execute(command)
}

func (c *dockerConnector) Connect(source *docker.Container, target *docker.Container) error {
	c.container = target.Id
	return nil
}

func (c *dockerConnector) Name() string {
	return "docker"
}

func (c *dockerConnector) Host() string {
	return c.container
}

func (c *dockerConnector) Args() []string {
	return []string{}
}

func (c *dockerConnector) Disconnect() error {
	return nil
}
