package connector

import (
	"fmt"
)

type sshConnector struct {
	host    string
	user    string
	keyPath string
}

func NewSsh(host string, user string, keyPath string) Connector {
	return &sshConnector{host: host, user: user, keyPath: keyPath}
}

func (c *sshConnector) Execute(executor Executor, playbook string) error {
	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-i", c.host + ",",
		"-l", c.host,
		"-vv",
		"--ssh-extra-args", fmt.Sprintf("-o StrictHostKeyChecking=no -o IdentityFile=%s", c.keyPath),
	}
	return executor.Execute(command)
}
