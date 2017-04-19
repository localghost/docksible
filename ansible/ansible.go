package ansible

import (
	"github.com/localghost/docksible/docker"
	"io/ioutil"
	"log"
	"os"
)

type Ansible struct {
	controller *docker.Container
	executor   Executor
	workDir    string
}

type PlayTarget struct {
	Container *docker.Container
	Connector Connector
	Groups    []string
}

func New(controller *docker.Container, executor Executor, workDir string) *Ansible {
	return &Ansible{controller: controller, executor: executor, workDir: workDir}
}

func (a *Ansible) Play(playbook string, target PlayTarget) error {
	inventoryFile := a.createInventory(target.Connector.Host(), target.Groups)
	defer os.Remove(inventoryFile)

	target.Connector.Connect(a.controller, target.Container)
	defer target.Connector.Disconnect()

	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-c", target.Connector.Name(),
		"-i", target.Connector.Host() + ",",
		"-l", target.Connector.Host(),
		"-vv",
	}
	command = append(command, target.Connector.Args()...)
	return a.executor.Execute(command)
}

func (a *Ansible) createInventory(host string, groups []string) string {
	file, err := ioutil.TempFile(os.TempDir(), "ansible-inventory")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// TODO: build the inventory

	return file.Name()
}
