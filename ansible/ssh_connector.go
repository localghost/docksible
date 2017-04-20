package ansible

import (
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/utils"
	"strings"
)

type sshConnector struct {
	host    string
	keyPath string
}

func NewSshConnector() Connector {
	return &sshConnector{}
}

func (c *sshConnector) Connect(source *docker.Container, target *docker.Container) error {
	sshKeys := utils.NewSSHKeyGenerator().GenerateInMemory()

	c.keyPath = "/tmp/id_rsa"
	source.CopyContentTo(c.keyPath, sshKeys.PrivateKey)

	target.CopyContentTo("/tmp", sshKeys.PublicKey)
	target.ExecAndWait("mkdir", "-p", "/root/.ssh")
	target.ExecAndWait("bash", "-c", "cat /tmp/id_rsa.pub >> /root/.ssh/authorized_keys")

	target.Exec("/usr/sbin/sshd", "-D")

	c.host = target.Inspect().NetworkSettings.IPAddress

	return nil
}

func (c *sshConnector) Name() string {
	return "ssh"
}

func (c *sshConnector) Host() string {
	return c.host
}

func (c *sshConnector) ExtraArgs() []string {
	sshExtraArgs := []string{
		"-o StrictHostKeyChecking=no",
		"-o IdentityFile=" + c.keyPath,
	}
	return []string{"--ssh-extra-args", strings.Join(sshExtraArgs, " ")}
}

func (c *sshConnector) Disconnect() error {
	// TODO: clean up keys and authorized_keys
	return nil
}
