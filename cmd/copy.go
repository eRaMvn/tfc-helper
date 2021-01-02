package cmd

import (
	"fmt"
	"os"
	"sync"
	"tfc-helper/helper"

	"github.com/spf13/cobra"
)

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Command to duplicate configuration between workspaces",
	Long: `Command to copy all variables existed in a workspace to another workspace. The destination workspace
can be the same organization or in a different organization

Examples:
- Copy all variables from workspace test1 to workspace test2 but ignore the variables that have the same name in test2. Both workspaces are in the same organization
tfc-help copy --src-ws test1 --dst-ws test2

OR if the source workspace and source organization have already been set with environment variables. We can do:
tfc-helper copy --dst-ws test2

- Copy all variables from workspace test1 to workspace test2 but overwrite the variables that have the same name in test2:
tfc-help copy --src-ws test1 --dst-ws test2 -r`,
	Run: func(cmd *cobra.Command, args []string) {
		// Try to get value from command line first then try the environment variable
		srcOrgName, _ := cmd.Flags().GetString("src-org")
		if srcOrgName == "" {
			// Try to get organization value from environment variable
			srcOrgName = os.Getenv(OrgVar)
			if srcOrgName == "" {
				fmt.Println("Please set the source organization with the --src-org flag or use TF_CLOUD_ORG_NAME environment variable")
				os.Exit(1)
			}
		}

		dstOrgName, _ := cmd.Flags().GetString("dst-org")

		// If dstOrgName is empty, we assume that the organization remains the same
		if dstOrgName == "" {
			dstOrgName = srcOrgName
		}

		// Try to get value from command line first then try the environment variable
		srcWsName, _ := cmd.Flags().GetString("src-ws")
		if srcWsName == "" {
			// Try to get workspace value from environment variable
			srcWsName = os.Getenv(WorkspaceVar)
			if srcWsName == "" {
				fmt.Println("Please set the source workspace with the --src-ws flag or use TF_CLOUD_WS_NAME environment variable")
				os.Exit(1)
			}
		}

		dstWsName, _ := cmd.Flags().GetString("dst-ws")
		shouldReplace, _ := cmd.Flags().GetBool("replace")

		srcWorkspaceID := helper.GetWorkspaceID(srcOrgName, srcWsName)
		dstWorkspaceID := helper.GetWorkspaceID(dstOrgName, dstWsName)

		variablesInDstWs := helper.ListAllVariables(dstWorkspaceID)

		var wg sync.WaitGroup
		for _, variable := range helper.ListAllVariables(srcWorkspaceID) {
			// Check if replace flag is set
			if shouldReplace {
				wg.Add(1)
				newVariable := helper.NewVariable{
					ID:          "",
					Key:         variable.Key,
					Value:       variable.Value,
					Description: variable.Description,
					Category:    variable.Category,
					HCL:         variable.HCL,
					Sensitive:   variable.Sensitive,
				}
				// If a variable already exists, delete it and create a new one
				if helper.CheckIfVariableExistInWs(variablesInDstWs, variable.Key) {
					// Try to get the variable ID
					variableToReplace, _ := helper.GetVar(dstWorkspaceID, variable.Key)
					newVariable.ID = variableToReplace.ID
					go helper.RecreateVariable(dstWorkspaceID, newVariable, &wg)
				} else {
					go helper.CreateVariable(dstWorkspaceID, newVariable, &wg)
				}
			} else {
				// If a variable already exists -- having the same name, skip
				if helper.CheckIfVariableExistInWs(variablesInDstWs, variable.Key) {
					continue
				} else {
					wg.Add(1)
					newVariable := helper.NewVariable{
						ID:          "",
						Key:         variable.Key,
						Value:       variable.Value,
						Description: variable.Description,
						Category:    variable.Category,
						HCL:         variable.HCL,
						Sensitive:   variable.Sensitive,
					}
					go helper.CreateVariable(dstWorkspaceID, newVariable, &wg)
				}
			}
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.PersistentFlags().String("src-org", "", "Specify the source organization")
	copyCmd.PersistentFlags().String("dst-org", "", "Specify the destination organization")
	copyCmd.PersistentFlags().String("src-ws", "", "Specify the source workspace")
	copyCmd.PersistentFlags().String("dst-ws", "", "Specify the destination workspace")
	_ = copyCmd.MarkPersistentFlagRequired("dst-ws")
	copyCmd.PersistentFlags().BoolP("replace", "r", false, "Specify whether to overwrite the existing variables or not")
}
