package ansible

import (
	"fmt"
	"strings"

	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/utils"
)

type sshConnector struct {
	host    string
	keyPath string
}

func NewSshConnector() Connector {
	return &sshConnector{}
}

func (c *sshConnector) Connect(source, target *docker.Container) error {
	c.installSSHKeys(source, target)

	targetInspect := target.Inspect()
	c.registerHostInContainer(source, targetInspect.NetworkSettings.IPAddress, targetInspect.Config.Hostname)

	c.host = targetInspect.Config.Hostname

	return nil
}

func (c *sshConnector) installSSHKeys(source, target *docker.Container) {
	sshKeys := utils.NewSSHKeyGenerator().GenerateInMemory()

	c.keyPath = "/tmp/id_rsa"
	source.CopyContentTo(c.keyPath, sshKeys.PrivateKey)
	source.ExecAndWait("chmod", "0400", c.keyPath)

	target.CopyContentTo("/tmp/id_rsa.pub", sshKeys.PublicKey)
	target.ExecAndWait("mkdir", "-p", "/root/.ssh")
	target.ExecAndWait("bash", "-c", "cat /tmp/id_rsa.pub >> /root/.ssh/authorized_keys")

	target.Exec("/usr/sbin/sshd", "-D")
}

func (c *sshConnector) registerHostInContainer(source *docker.Container, ipAddress, hostname string) {
	command := fmt.Sprintf(`echo "%s %s" >> /etc/hosts`, ipAddress, hostname)
	source.ExecAndWait("bash", "-c", command)
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
	return []string{"--ssh-extra-args", fmt.Sprintf("'%s'", strings.Join(sshExtraArgs, " "))}
}

func (c *sshConnector) Disconnect() error {
	// TODO: clean up keys and authorized_keys
	return nil
}
