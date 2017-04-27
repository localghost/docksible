package cmd

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/localghost/docksible/ansible"
	"github.com/localghost/docksible/docker"
	"github.com/localghost/docksible/provisioner"
	"github.com/localghost/docksible/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type rootFlags struct {
	ansibleDir       string
	playbookPath     string
	inventoryGroups  []string
	extraArgs        []string
	builderImage     string
	resultImage      string
	ansibleConnector string `choices:"docker-exec,ssh"`
}

func getChoices(flags *rootFlags, fieldName string) []string {
	field, ok := reflect.TypeOf(*flags).FieldByName(fieldName)
	if !ok {
		panic(fmt.Sprintf("Field %s not found", fieldName))
	}
	choices := field.Tag.Get(`choices`)
	if choices == "" {
		return []string{}
	}
	return strings.Split(choices, ",")
}

func getDefaultAnsibleDir() string {
	defaultAnsibleDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return defaultAnsibleDir
}

func CreateRootCommand() *cobra.Command {
	flags := rootFlags{}

	ansibleConnectorChoices := getChoices(&flags, `ansibleConnector`)

	cmd := &cobra.Command{
		Use: "docksible [flags] image playbook",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Please provide path to playbook to execute and image to provision.")
			}
			if !utils.InStringSlice(flags.ansibleConnector, ansibleConnectorChoices) {
				return fmt.Errorf("%s is not a supported ansible connector", flags.ansibleConnector)
			}
			var err error
			if args[1], err = filepath.Abs(args[1]); err != nil {
				log.Fatal(err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(args[0], args[1], args[2:], &flags); err != nil {
				fmt.Println(err)
			}
		},
	}

	cmd.Flags().StringVarP(&flags.ansibleDir, "ansible-dir", "a", getDefaultAnsibleDir(), "Path to ansible directory.")
	cmd.Flags().StringSliceVarP(&flags.inventoryGroups, "inventory-group", "g", []string{}, "Ansible group the provisioned container should belong to.")
	cmd.Flags().StringVarP(
		&flags.builderImage, "builder-image", "b", "docksible/builder:latest",
		"Docker image for the builder container. See documentation for its requirements.",
	)
	cmd.Flags().StringVarP(&flags.resultImage, "result-image", "r", "", "Name of the resulting docker image.")
	cmd.Flags().StringVarP(
		&flags.ansibleConnector, "ansible-connector", "c", ansibleConnectorChoices[0],
		fmt.Sprintf("Ansible connector type to use (choices: %s)", strings.Join(ansibleConnectorChoices, ", ")),
	)

	return cmd
}

func run(image, playbook string, ansibleExtraArgs []string, flags *rootFlags) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	provisioner, err := provisioner.NewProvisioner(flags.builderImage, cli).Run(flags.ansibleDir, playbook)
	if err != nil {
		return err
	}
	defer provisioner.StopAndRemove()

	provisioned, err := runProvisioned(image, cli)
	if err != nil {
		return err
	}
	defer provisioned.StopAndRemove()

	ans := ansible.New(provisioner, flags.ansibleDir)
	target := ansible.PlayTarget{
		Container: provisioned,
		Connector: ansible.CreateConnector(flags.ansibleConnector),
		Groups:    flags.inventoryGroups,
	}
	err = ans.Play(playbook, target, ansibleExtraArgs)
	if err != nil {
		return err
	}

	imageId, err := provisioned.Commit(flags.resultImage, "bash")
	if err != nil {
		return err
	}
	fmt.Printf("Image %s(%s) built successfully.\n", flags.resultImage, imageId)

	return nil
}

func runProvisioned(image string, cli *client.Client) (*docker.Container, error) {
	config := &container.Config{
		Cmd:        []string{"tail", "-f", "/dev/null"},
		Image:      image,
		StopSignal: "SIGKILL",
	}
	return docker.NewContainer("", config, nil, nil, cli)
}
