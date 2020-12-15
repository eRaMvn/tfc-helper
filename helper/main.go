package helper

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

var client *tfe.Client
var err error
var ctx context.Context

type NewVariable struct {
	ID          string           `jsonapi:"primary,vars"`
	Key         string           `jsonapi:"attr,key"`
	Value       string           `jsonapi:"attr,value"`
	Description string           `jsonapi:"attr,description"`
	Category    tfe.CategoryType `jsonapi:"attr,category"`
	HCL         bool             `jsonapi:"attr,hcl"`
	Sensitive   bool             `jsonapi:"attr,sensitive"`
}

type Config struct {
	Credentials CredentialConfig `hcl:"credentials,block"`
}

type CredentialConfig struct {
	App   string `hcl:"app,label"`
	Token string `hcl:"token"`
}

// GetCommandValues gets variables from the command line
func GetCommandValues(values []string) map[string]string {
	valueToSend := make(map[string]string)

	for _, e := range values {
		pair := strings.SplitN(e, "=", 2)
		// Check when only the variable name is supplied
		if len(pair) == 1 {
			valueToSend[pair[0]] = ""
		} else {
			valueToSend[pair[0]] = pair[1]
		}
	}
	return valueToSend
}

// GetTFValues gets all Terraform environment variables
// Takes a parameter to get the value returned as Terraform or Env variables
func GetTFValues(isTVar bool) map[string]string {
	valueToSend := make(map[string]string)

	//  Get all environment variables and values
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "TF_VAR_") {
			if isTVar {
				valueToSend[pair[0]] = pair[1]
			} else {
				varName := strings.SplitN(pair[0], "TF_VAR_", 2)
				valueToSend[varName[1]] = pair[1]
			}
		}
	}
	return valueToSend
}

// ListAllWorkspaces lists all workspaces in the organization
func ListAllWorkspaces(organizationName string) []*tfe.Workspace {
	workspaceList, err := client.Workspaces.List(ctx, organizationName, tfe.WorkspaceListOptions{})
	if err != nil {
		fmt.Println("Organization not found or incorrect! Please set the environment variable or the flag value again")
		os.Exit(1)
	}
	return *&workspaceList.Items
}

// GetWorkspaceID gets the workspace id in the list of workspace in the organization
func GetWorkspaceID(organizationName string, workspaceName string) string {
	workspaceID := ""
	for _, workspace := range ListAllWorkspaces(organizationName) {
		if workspace.Name == workspaceName {
			workspaceID = workspace.ID
		}
	}
	if workspaceID == "" {
		fmt.Println("Workspace name not found or incorrect! Please set the environment variable or the flag value again")
		os.Exit(1)
	}
	return workspaceID
}

// ListAllVariables list all variables (terraform and environment variables) in the workspace
func ListAllVariables(workspaceID string) []*tfe.Variable {
	variableList, err := client.Variables.List(ctx, workspaceID, tfe.VariableListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return *&variableList.Items
}

// GetVar gets the variable that matches the variable name in the list of variable
func GetVar(workspaceID string, varName string) (variable *tfe.Variable, err error) {
	for _, variable := range ListAllVariables(workspaceID) {
		if variable.Key == varName {
			return variable, nil
		}
	}
	return nil, errors.New("variable not found")
}

// CreateVariable creates a variable
func CreateVariable(workspaceID string, newVariable NewVariable) {
	_, err := client.Variables.Create(ctx, workspaceID, tfe.VariableCreateOptions{
		Key:         tfe.String(newVariable.Key),
		Value:       tfe.String(newVariable.Value),
		Description: tfe.String(newVariable.Description),
		Category:    tfe.Category(newVariable.Category),
		HCL:         tfe.Bool(newVariable.HCL),
		Sensitive:   tfe.Bool(newVariable.Sensitive),
	})
	if err != nil {
		log.Fatal(err)
	}
}

// UpdateVariable updates a variable given the variable id
func UpdateVariable(workspaceID string, newVariable NewVariable) {
	_, err := client.Variables.Update(ctx, workspaceID, newVariable.ID, tfe.VariableUpdateOptions{
		Value:       tfe.String(newVariable.Value),
		Description: tfe.String(newVariable.Description),
		HCL:         tfe.Bool(newVariable.HCL),
		Sensitive:   tfe.Bool(newVariable.Sensitive),
	})

	if err != nil {
		log.Fatal(err)
	}
}

// DeleteVariable deletes a variable or all variables
func DeleteVariable(workspaceID string, variableID string, all bool) {
	if all {
		for _, variable := range ListAllVariables(workspaceID) {
			client.Variables.Delete(ctx, workspaceID, variable.ID)
		}
	} else {
		client.Variables.Delete(ctx, workspaceID, variableID)
	}
}

// RecreateVariable deletes a variable and create it again
func RecreateVariable(workspaceID string, variable NewVariable) {
	DeleteVariable(workspaceID, variable.ID, false)
	CreateVariable(workspaceID, variable)
}

func init() {
	// Get token from terraformrc from Linux
	// TODO: implement this in Windows
	homeDir, _ := os.UserHomeDir()
	var terraformCloudToken string
	content, fileReadingError := ioutil.ReadFile(fmt.Sprintf("%s/.terraformrc", homeDir))
	if fileReadingError != nil {
		// Read from environment variable TF_CLOUD_TOKEN
		terraformCloudToken = os.Getenv("TF_CLOUD_TOKEN")
	} else {
		// Structure of terraformrc file
		var terraformrcConfig Config
		decodeErr := hclsimple.Decode(
			"somefile.hcl", []byte(content),
			nil, &terraformrcConfig,
		)
		if decodeErr != nil {
			log.Fatalf("Failed to load configuration: %s", decodeErr)
		}
		terraformCloudToken = terraformrcConfig.Credentials.Token
	}

	if terraformCloudToken == "" {
		fmt.Println("Token not available!")
		return
	}

	tFconfig := &tfe.Config{
		Token: terraformCloudToken,
	}

	client, err = tfe.NewClient(tFconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx = context.Background()
}