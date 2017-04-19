package ansible

import (
	"strings"
)

type sshConnector struct {
	host    string
	user    string
	keyPath string
}

func NewSshConnector(host string, user string, keyPath string) Connector {
	return &sshConnector{host: host, user: user, keyPath: keyPath}
}

func (c *sshConnector) Execute(executor Executor, playbook string) error {
	sshExtraArgs := []string{
		"-o StrictHostKeyChecking=no",
		"-o IdentityFile=" + c.keyPath,
	}
	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-i", c.host + ",",
		"-l", c.host,
		"-vv",
		"--ssh-extra-args", strings.Join(sshExtraArgs, " "),
	}
	return executor.Execute(command)
}
