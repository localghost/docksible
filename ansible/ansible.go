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
	if err := target.Connector.Connect(a.controller, target.Container); err != nil {
		return err
	}
	defer target.Connector.Disconnect()

	// Host() is valid only after connection between containers have been established.
	inventory, err := a.createInventory(target.Connector.Host(), target.Groups)
	if err != nil {
		return err
	}
	inventoryPath := "/tmp/docksible-inventory"
	if a.workDir != "" {
		inventoryPath = filepath.Join(a.workDir, "docksible-inventory")
	}
	if err := a.controller.CopyContentTo(inventoryPath, inventory); err != nil {
		return err
	}
	if _, err := a.controller.ExecAndWait("chmod", "0666", inventoryPath); err != nil {
		return nil
	}
	// FIXME: check error in defer (use hashicorp/go-mutlierror?)
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

func (a *Ansible) createInventory(host string, groups []string) (*bytes.Buffer, error) {
	inventory := new(bytes.Buffer)

	if _, err := inventory.WriteString(fmt.Sprintf("%s ansible_user=root\n", host)); err != nil {
		return nil, err
	}

	cfg := ini.Empty()
	for _, group := range groups {
		if _, err := cfg.Section(group).NewBooleanKey(host); err != nil {
			return nil, err
		}
	}
	if _, err := cfg.WriteTo(inventory); err != nil {
		return nil, err
	}

	return inventory, nil
}
