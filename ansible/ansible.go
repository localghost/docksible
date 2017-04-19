package ansible

import (
	"github.com/localghost/docksible/docker"
	"io/ioutil"
	"log"
	"os"
)

type Ansible struct {
	controller *docker.Container
	workDir    string
}

type PlayTarget struct {
	container *docker.Container
	connector Connector
	groups    []string
}

func New(controller *docker.Container, workDir string) *Ansible {
	return &Ansible{controller: controller, workDir: workDir}
}

func (a *Ansible) Play(playbook string, target PlayTarget) {
	// TODO: retrieve host from connector
	inventoryFile := a.createInventory("", target.groups)
	defer os.Remove(inventoryFile)
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
