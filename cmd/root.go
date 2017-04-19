package cmd

import (
	"fmt"
	"github.com/localghost/docksible/builder"
	"github.com/localghost/docksible/product"
	"github.com/spf13/cobra"
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

func inSlice(needle string, haystack []string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

func CreateRootCommand() *cobra.Command {
	flags := rootFlags{}

	ansibleConnectorChoices := getChoices(&flags, `ansibleConnector`)

	cmd := &cobra.Command{
		Use: "docksible [flags] image playbook",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("Please provide path to playbook to execute and image to provision.")
			}
			if !inSlice(flags.ansibleConnector, ansibleConnectorChoices) {
				return fmt.Errorf("%s is not a supported ansible connector", flags.ansibleConnector)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			run(args[0], args[1], &flags)
		},
	}
	cmd.Flags().StringVarP(&flags.ansibleDir, "ansible-dir", "a", "", "Path to ansible directory.")
	cmd.Flags().StringSliceVarP(&flags.inventoryGroups, "inventory-groups", "g", []string{}, "Ansible groups the provisioned container should belong to.")
	cmd.Flags().StringSliceVarP(&flags.extraArgs, "extra-args", "x", []string{}, "Extra arguments passed to ansible.")
	cmd.Flags().StringVarP(&flags.builderImage, "builder-image", "b", "docksible/builder", "Docker image for the builder container. It needs to have ansible and ssh client built in.")
	cmd.Flags().StringVarP(&flags.resultImage, "result-image", "r", "", "Name of the resulting docker image.")
	cmd.Flags().StringVarP(
		&flags.ansibleConnector, "ansible-connector", "c", ansibleConnectorChoices[0],
		fmt.Sprintf("Ansible connector type to use (choices: %s)", strings.Join(ansibleConnectorChoices, ", ")),
	)

	return cmd
}

func run(image, playbook string, flags *rootFlags) {
	provisionOptions := &builder.ProvisionOptions{
		AnsibleDir:      flags.ansibleDir,
		PlaybookPath:    playbook,
		InventoryGroups: flags.inventoryGroups,
	}
	b := builder.New(flags.builderImage)

	provisioned := product.New(image).Run()
	defer provisioned.StopAndRemove()

	b.ProvisionContainer(provisioned, provisionOptions)

	imageId := provisioned.Commit(flags.resultImage, "bash")
	fmt.Printf("Image %s built successfully.\n", imageId)
}
