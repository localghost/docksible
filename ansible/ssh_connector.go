package ansible

import (
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/utils"
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

func (c *sshConnector) Connect(source *docker.Container, target *docker.Container) error {
	sshKeys := utils.NewSSHKeyGenerator().GenerateInMemory()

	bc := utils.NewInMemoryArchive()
	bc.AddBytes("id_rsa", sshKeys.PrivateKey.Bytes())
	privateKey := bc.Close()
	source.CopyTo("/tmp", privateKey)
	c.keyPath = "/tmp/id_rsa"

	bc = utils.NewInMemoryArchive()
	bc.AddBytes("id_rsa.pub", sshKeys.PublicKey.Bytes())
	publicKey := bc.Close()
	target.CopyTo("/tmp", publicKey)
	target.ExecAndWait("mkdir", "-p", "/root/.ssh")
	target.ExecAndWait("bash", "-c", "cat /tmp/id_rsa.pub >> /root/.ssh/authorized_keys")

	target.Exec("/usr/sbin/sshd", "-D")

	response := target.Inspect()
	c.host = response.NetworkSettings.IPAddress

	return nil
}

func (c *sshConnector) Name() string {
	return "ssh"
}

func (c *sshConnector) Host() string {
	return c.host
}

func (c *sshConnector) Args() []string {
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
