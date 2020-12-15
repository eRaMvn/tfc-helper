/*
Copyright © 2020 eRaMvn pdtvnhcm@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
		keyPairs, _ := cmd.Flags().GetStringSlice("var")
		workspaceName, _ := cmd.Flags().GetString("workspace")
		organizationName, _ := cmd.Flags().GetString("organization")
		allVar, _ := cmd.Flags().GetBool("all")

		workspaceID := helper.GetWorkspaceID(organizationName, workspaceName)

		if allVar {
			// Delete all variables in the workspace
			helper.DeleteVariable(workspaceID, "", allVar)
		} else {
			valueToSend := helper.GetCommandValues(keyPairs)
			// Loop through all values passed from the command line
			for newVariableName := range valueToSend {
				// Try to get the variable ID
				variable, error := helper.GetVar(workspaceID, newVariableName)

				/* If error, meaning the variable does not exist, print out error message */
				if error != nil {
					fmt.Println("Variable does not exist. Please use tfc-help delete command to delete the variable!")
					os.Exit(1)
					// When variable already exists, proceed to delete the variable
				} else {
					helper.DeleteVariable(workspaceID, variable.ID, false)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().BoolP("all", "a", false, "Specify whether to delete all variables")
}
