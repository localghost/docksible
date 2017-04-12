package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"log"
	"fmt"
)

type topLevelFlags struct {
	ansibleDir string
	playbookPath string
}

func getCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func printUsageError(cmd *cobra.Command, message string) {
	fmt.Printf("%s\n\n%s", message, cmd.UsageString())
}

func CreateRootCommand() *cobra.Command {
	flags := topLevelFlags{}
	cmd := &cobra.Command{
		Use: "docksible [flags] playbook",
		Run: func (cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				printUsageError(cmd, "Please provide path to playbook to execute.")
			}
		},
	}
	cmd.Flags().StringVarP(&flags.ansibleDir, "ansible-dir", "a", getCwd(), "Path to ansible directory.")

	return cmd
}

