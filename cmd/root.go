/*
Package cmd includes all arguments and flags for the tool
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

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var workspaceName string
var organizationName string

// WorkspaceVar sets the workspace variable name the tool will look for
const WorkspaceVar = "TF_CLOUD_WS_NAME"

// OrgVar sets the organization variable name the tool will look for
const OrgVar = "TF_CLOUD_ORG_NAME"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tfc-helper",
	Short: "A tool to help prepare workspace for Terraform Cloud",
	Long: `Using Terraform requires the variables to be set up along with the
credentials for the provider in order to run. tfc-help was created to help
automate that process.
Example:
tfc-help update --var some_variable=some_value -d "This variable is just an example" -w ws-K33Rp -o big-corp
tfc-help delete --var some_variable -w ws-K33Rp -o big-corp

For more information, please use:
tfc-helper [create/update/delete] -h
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("workspace", "w", "", "Specify the name of the workspace")
	rootCmd.PersistentFlags().StringP("organization", "o", "", "Specify the name of the organization")
	rootCmd.PersistentFlags().StringSlice("var", []string{}, `Key value pair to put in terraform cloud.
This flag can be set multiple times and can multiple values with comma separated`)
	rootCmd.PersistentFlags().StringP("description", "d", "", "Specify the description for the variable")
	rootCmd.PersistentFlags().BoolP("terraform", "t", false, "Specify whether values are terraform variable")
	rootCmd.PersistentFlags().Bool("hcl", false, "Specify whether the values are in HCL format")
	rootCmd.PersistentFlags().BoolP("sensitive", "s", false, "Specify whether the values are sensitive")
	rootCmd.PersistentFlags().BoolP("keep", "k", false, "Specify whether to keep the original value of the variable")
	rootCmd.PersistentFlags().BoolP("env", "e", false, "Specify whether to grab the environment variables starting with 'TF_VAR_' from the host")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tfc-helper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tfc-helper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
