package ansible

import (
	"bytes"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/localghost/docksible/docker"
	"log"
	"os"
	"path/filepath"
)

type Ansible struct {
	controller *docker.Container
	workDir    string
}

type PlayTarget struct {
	Container *docker.Container
	Connector Connector
	Groups    []string
}

// TODO Pass stream(s) used in ExecAndOutput
func New(controller *docker.Container, workDir string) *Ansible {
	return &Ansible{controller: controller, workDir: workDir}
}

func (a *Ansible) Play(playbook string, target PlayTarget, extraArgs []string) error {
	target.Connector.Connect(a.controller, target.Container)
	defer target.Connector.Disconnect()

	// Host() is valid only after connection between containers have been established.
	inventory := a.createInventory(target.Connector.Host(), target.Groups)
	inventoryPath := "/tmp/docksible-inventory"
	if a.workDir != "" {
		inventoryPath = filepath.Join(a.workDir, "docksible-inventory")
	}
	a.controller.CopyContentTo(inventoryPath, inventory)
	defer a.controller.ExecAndWait("rm", "-rf", inventoryPath)

	command := []string{
		"/usr/bin/ansible-playbook",
		playbook,
		"-c", target.Connector.Name(),
		"-i", inventoryPath,
		"-vv",
	}
	command = append(command, target.Connector.ExtraArgs()...)
	command = append(command, extraArgs...)
	code, err := a.controller.ExecAndOutput(os.Stdout, os.Stderr, command...)
	if err != nil {
		log.Fatal(err)
	}
	if code != 0 {
		return fmt.Errorf("Command failed with %d", code)
	}
	return nil
}

func (a *Ansible) createInventory(host string, groups []string) *bytes.Buffer {
	inventory := new(bytes.Buffer)

	cfg := ini.Empty()
	_, _ = cfg.Section("").NewBooleanKey(host) // TODO check error
	for _, group := range groups {
		_, _ = cfg.Section(group).NewBooleanKey(host) // TODO check error
	}
	cfg.WriteTo(inventory)

	return inventory
}
