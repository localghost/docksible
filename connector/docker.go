package connector

type dockerConnector struct {
	container string
	user      string
}

func NewDocker(container string, user string) Connector {
	return &dockerConnector{container: container, user: user}
}

func (c *dockerConnector) Execute(executor Executor, playbook string) error {
	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-c", "docker",
		"-i", "localhost,",
		"-l", "localhost",
		"-e ansible_host=" + c.container,
		"-e ansible_user=" + c.user,
		"-vv",
	}
	return executor.Execute(command)
}
