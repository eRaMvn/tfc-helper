package cmd

import (
	"fmt"
	"os"
	"tfc-helper/helper"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Command to create/update variables in a TF workspace",
	Long: `Command used to create variables if variables don't exist in
TF workspace. If the variables do exist, the command can update the
current variables.
In use cases when we need to update a sensitive variable to non-sensitive or when we
need to change all variables from being terraform variables to environment variables and vice versa.
We can delete the current variables and create new ones with -r flag.

Examples:
- Create/Update a single variable:
tfc-help update --var some_variable=some_value -d "This variable is just an example" -w ws-K33Rp -o big-corp

- Update everything about variable except for its value and description. Flag -k is to make sure existing values are kept.
Without -k, value of a variable will be empty/blank
tfc-help update --var some_variable -k -r -w ws-K33Rp -o big-corp

- Create/Update all variables in a workspace with variables in the current environment.
Flag -r is to avoid issue when we have existing variable that is in one category (terraform or env, encrypt or not encrypt)
but the updated value is another category:
tfc-help update --env -r -w ws-K33Rp -o big-corp

- Create/Update only new variables found in the current workspace:
tfc-help update --env -w ws-K33Rp -o big-corp`,
	Run: func(cmd *cobra.Command, args []string) {
		// Try to get workspace value from environment variable
		wsName := os.Getenv(WorkspaceVar)
		if wsName != "" {
			workspaceName = wsName
		} else {
			workspaceName, _ = cmd.Flags().GetString("workspace")
		}

		// Try to get organization value from environment variable
		orgName := os.Getenv(OrgVar)

		if orgName != "" {
			organizationName = orgName
		} else {
			organizationName, _ = cmd.Flags().GetString("organization")
		}

		keyPairs, _ := cmd.Flags().GetStringSlice("var")
		variableDescription, _ := cmd.Flags().GetString("description")
		isTVar, _ := cmd.Flags().GetBool("terraform")
		hcl, _ := cmd.Flags().GetBool("hcl")
		sensitive, _ := cmd.Flags().GetBool("sensitive")
		shouldReplace, _ := cmd.Flags().GetBool("replace")
		keepValue, _ := cmd.Flags().GetBool("keep")
		env, _ := cmd.Flags().GetBool("env")

		// Get the correct category type
		category := tfe.CategoryEnv
		if isTVar {
			category = tfe.CategoryTerraform
		}

		workspaceID := helper.GetWorkspaceID(organizationName, workspaceName)

		var valueToSend map[string]string

		if env && len(keyPairs) > 0 {
			valueToSend = helper.GetTFValues(isTVar)
			commandValues := helper.GetCommandValues(keyPairs)
			for key := range valueToSend {
				for commandKey, commandValue := range commandValues {
					// Check if there are duplicated key, values in environment win
					if key == commandKey {
						continue
					} else {
						valueToSend[commandKey] = commandValue
					}
				}

			}
			// If only flag -e is set, then grab all variables in the current environment
		} else if env {
			valueToSend = helper.GetTFValues(isTVar)
			// If only flag -v is set, then grab all variables in the command line
		} else {
			valueToSend = helper.GetCommandValues(keyPairs)
		}

		// Loop through all values passed from the command line
		for newVariableName, newVariableValue := range valueToSend {
			// Try to get the variable ID
			variable, error := helper.GetVar(workspaceID, newVariableName)

			/* If error, meaning the variable does not exist, proceed to create one based on
			current value */
			if error != nil {
				// newVariable stores the value given in the command
				newVariable := helper.NewVariable{
					ID:          "",
					Key:         newVariableName,
					Value:       newVariableValue,
					Description: variableDescription,
					Category:    category,
					HCL:         hcl,
					Sensitive:   sensitive,
				}

				helper.CreateVariable(workspaceID, newVariable)
				// When variable already exists, proceed to update the variable
			} else {
				// originalVariable keeps the value and description given to a variable but changes the hcl and encryption type
				originalVariable := helper.NewVariable{
					ID:          variable.ID,
					Key:         newVariableName,
					Value:       variable.Value,
					Description: variable.Description,
					Category:    category,
					HCL:         hcl,
					Sensitive:   sensitive,
				}

				// newVariable stores the value given in the command
				newVariable := helper.NewVariable{
					ID:          variable.ID,
					Key:         newVariableName,
					Value:       newVariableValue,
					Description: variableDescription,
					Category:    category,
					HCL:         hcl,
					Sensitive:   sensitive,
				}

				// If -r flag is set, proceed to recreate the variable
				if shouldReplace {
					if keepValue {
						helper.RecreateVariable(workspaceID, originalVariable)
					} else {
						helper.RecreateVariable(workspaceID, newVariable)
					}

				} else {
					/* If change from sensitive to non-sensitive or from terraform to env and vice versa,
					print out message to use -r */
					if variable.Category != category || variable.Sensitive != sensitive {
						sampleCommand1 := fmt.Sprintf("tfc-help update --var %s -k -r -w %s -o %s", newVariableName, workspaceName, organizationName)
						fmt.Println(`One of the variables cannot be updated. Please use -r flag to recreate the variable.
Changing from one type to another or marking a variable from sensitive to non-sensitive requires variable recreation.
Also, if you want to keep the existing variable value and description, please use -k flag
Example commands:
`, sampleCommand1)
						os.Exit(0)
					}

					// If -k flag is set, keep the original value of the variable
					if keepValue {
						helper.UpdateVariable(workspaceID, originalVariable)
					} else {
						// TODO: Create a way to keep the description the same but the value can be different
						helper.UpdateVariable(workspaceID, newVariable)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().BoolP("replace", "r", false, "Specify whether to replace the existing variable or not")
}
