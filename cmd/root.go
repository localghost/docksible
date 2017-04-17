package cmd

import (
	"fmt"
	"github.com/localghost/docksible/builder"
	"github.com/localghost/docksible/product"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	ansibleDir      string
	playbookPath    string
	inventoryGroups []string
	extraArgs       []string
	builderImage    string
	resultImage     string
}

func CreateRootCommand() *cobra.Command {
	flags := rootFlags{}
	cmd := &cobra.Command{
		Use: "docksible [flags] image playbook",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("Please provide path to playbook to execute and image to provision.")
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
	defer provisioned.Remove()
	b.ProvisionContainer(provisioned, provisionOptions)

	imageId := provisioned.Commit(flags.resultImage, "bash")
	fmt.Printf("Image %s built successfully.\n", imageId)
}
