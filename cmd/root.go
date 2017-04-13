package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/localghost/docksible/product"
)

type rootFlags struct {
	ansibleDir string
	playbookPath string
	inventoryGroups []string
	extraArgs []string
	baseImage string
	builderImage string
}

func printUsageError(cmd *cobra.Command, message string) {
	fmt.Printf("%s\n\n%s", message, cmd.UsageString())
}

func CreateRootCommand() *cobra.Command {
	flags := rootFlags{}
	cmd := &cobra.Command{
		Use: "docksible [flags] image playbook",
		Run: func (cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				printUsageError(cmd, "Please provide path to playbook to execute and image to provision.")
			}
			run(args[0], args[1], &flags)
		},
	}
	cmd.Flags().StringVarP(&flags.ansibleDir, "ansible-dir", "a", "", "Path to ansible directory.")
	cmd.Flags().StringSliceVarP(&flags.inventoryGroups, "inventory-groups", "g", []string{}, "Ansible groups the provisioned container should belong to.")
	cmd.Flags().StringSliceVarP(&flags.extraArgs, "extra-args", "x", []string{}, "Extra arguments passed to ansible.")
	cmd.Flags().StringVarP(&flags.builderImage, "builder-image", "b", "docksible/builder", "Docker image for the builder container. It needs to have ansible and ssh client built in.")

	return cmd
}

func run(image, playbook string, flags *rootFlags) {
	builder := builder.New()
	builder.Bootstrap()
	builder.ProvisionContainer(product.New().Run())
}
