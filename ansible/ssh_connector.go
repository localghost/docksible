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
	if err := c.installSSHKeys(source, target); err != nil {
		return err
	}

	targetInspect, err := target.Inspect()
	if err != nil {
		return err
	}
	if err := c.registerHostInContainer(source, targetInspect.NetworkSettings.IPAddress, targetInspect.Config.Hostname); err != nil {
		return err
	}

	c.host = targetInspect.Config.Hostname

	return nil
}

func (c *sshConnector) installSSHKeys(source, target *docker.Container) error {
	sshKeys, err := utils.NewSSHKeyGenerator().GenerateInMemory()
	if err != nil {
		return err
	}

	c.keyPath = "/tmp/id_rsa"
	if err := source.CopyContentTo(c.keyPath, sshKeys.PrivateKey); err != nil {
		return err
	}
	// FIXME: check exit code
	if _, err := source.ExecAndWait("chmod", "0400", c.keyPath); err != nil {
		return err
	}

	if err := target.CopyContentTo("/tmp/id_rsa.pub", sshKeys.PublicKey); err != nil {
		return err
	}
	if _, err := target.ExecAndWait("mkdir", "-p", "/root/.ssh"); err != nil {
		return err
	}
	if _, err := target.ExecAndWait("bash", "-c", "cat /tmp/id_rsa.pub >> /root/.ssh/authorized_keys"); err != nil {
		return err
	}

	if err := target.Exec("/usr/sbin/sshd", "-D"); err != nil {
		return err
	}

	return nil
}

func (c *sshConnector) registerHostInContainer(source *docker.Container, ipAddress, hostname string) error {
	command := fmt.Sprintf(`echo "%s %s" >> /etc/hosts`, ipAddress, hostname)
	if _, err := source.ExecAndWait("bash", "-c", command); err != nil {
		return err
	}
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
	return []string{"--ssh-extra-args", fmt.Sprintf("'%s'", strings.Join(sshExtraArgs, " "))}
}

func (c *sshConnector) Disconnect() error {
	// TODO: clean up keys and authorized_keys
	return nil
}
