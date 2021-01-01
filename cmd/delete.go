package cmd

import (
	"fmt"
	"os"
	"tfc-helper/helper"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Command to delete variables in a TF workspace",
	Long: `Command used to delete variable(s). Use -a flag to
delete all variables in the current workspace. Need to set appropriate
flag to delete terraform variables or environment variables with -e.
Examples:
tfc-help delete --var some_variable=some_value -w ws-K33Rp -o big-corp
tfc-help delete --var some_variable -w ws-K33Rp -o big-corp
tfc-help delete -a -w ws-K33Rp -o big-corp`,
	Run: func(cmd *cobra.Command, args []string) {
		// Try to get workspace value from environment variable
		// var workspaceName string
		wsName := os.Getenv(WorkspaceVar)
		if wsName != "" {
			workspaceName = wsName
		} else {
			workspaceName, _ = cmd.Flags().GetString("workspace")
		}

		// Try to get organization value from environment variable
		// var organizationName string
		orgName := os.Getenv(OrgVar)

		if orgName != "" {
			organizationName = orgName
		} else {
			organizationName, _ = cmd.Flags().GetString("organization")
		}

		keyPairs, _ := cmd.Flags().GetStringSlice("var")
		allVar, _ := cmd.Flags().GetBool("all")

		workspaceID := helper.GetWorkspaceID(organizationName, workspaceName)

		if allVar {
			emptyID := ""
			// Delete all variables in the workspace
			helper.DeleteVariables(workspaceID, emptyID, allVar)
		} else {
			valueToSend := helper.GetCommandValues(keyPairs)
			// Loop through all values passed from the command line
			for newVariableName := range valueToSend {
				// Try to get the variable ID
				variable, error := helper.GetVar(workspaceID, newVariableName)

				/* If error, meaning the variable does not exist, print out error message */
				if error != nil {
					fmt.Println("Variable does not exist. Cannot delete the variable!")
					os.Exit(1)
					// When variable already exists, proceed to delete the variable
				} else {
					helper.DeleteVariables(workspaceID, variable.ID, false)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().BoolP("all", "a", false, "Specify whether to delete all variables")
}
