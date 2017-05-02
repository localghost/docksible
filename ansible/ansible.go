package ansible

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	"github.com/localghost/docksible/docker"
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

	ansibleCommand := []string{
		"ansible-playbook",
		playbook,
		"-c", target.Connector.Name(),
		"-i", inventoryPath,
	}
	ansibleCommand = append(ansibleCommand, target.Connector.ExtraArgs()...)
	ansibleCommand = append(ansibleCommand, extraArgs...)

	command := []string{
		"sh", "-c", fmt.Sprintf("cd %s && %s", a.workDir, strings.Join(ansibleCommand, " ")),
	}
	code, err := a.controller.ExecAndOutput(os.Stdout, os.Stderr, command...)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("Provisioning failed [%d].", code)
	}
	return nil
}

func (a *Ansible) createInventory(host string, groups []string) *bytes.Buffer {
	inventory := new(bytes.Buffer)

	inventory.WriteString(fmt.Sprintf("%s ansible_user=root\n", host))

	cfg := ini.Empty()
	for _, group := range groups {
		_, _ = cfg.Section(group).NewBooleanKey(host) // TODO check error
	}
	cfg.WriteTo(inventory)

	return inventory
}
