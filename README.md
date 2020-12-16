# `tfc-helper`

tfc-helper is a CLI tool to help with updating variables in Terraform Cloud or Terraform Enterprise. The tool can be used within pipeline to easily update the variables' values in Terraform Cloud, or it can be used to make the task of testing terraform deployment using Terraform Cloud easier

## Pre-requisites

You will need:

1. Linux Environment (Windows Environment is under development)
2. Terraform Cloud API key (which can be found under [setting](https://app.terraform.io/app/settings/tokens))
3. Configure `terraformrc` file according to [this](https://www.terraform.io/docs/commands/cli-config.html). `terraformrc` file should be under your home directory.

   1. Instead of using `terraformrc` file. You can use the following environment variable with
`
export TF_CLOUD_TOKEN=[your_token]
`

1. Organization name where workspace is created. Organization name can be passed with -o flag (see samples below) or through env variable `TF_CLOUD_ORG_NAME`
2. Workspace name where variables are created. Workspace name can be passed with -w flag (see samples below) or through env variable `TF_CLOUD_WS_NAME`

## Installing

`git clone https://github.com/eRaMvn/tfc-helper.git`

Build executable

```bash
#!/bin/bash
go mod tidy
go build
```

## Usage

There are two main commands:

- update: To create/update variables in Terraform Cloud
- delete: To delete variables in Terraform Cloud

By default, the tool assumes that the variable will be environment variable. It will not marked as sensitive or as HCL value.

If you want any of the capability, please use:

- Flag -s to mark variable as sensitive
- Flag --hcl to mark value of the variable as hcl value
- Flag -t to mark the variable as terraform variable

## Example Commands

1. Create variable(s) with default flags (not HCL and not sensitive value) assuming the variable you create does not exist already. The same command can be used to update the variable(s)

`
tfc-help update --var some_variable=some_value -o sample-org -w sample-workspace
`

If organization and workspace variables are already set with

```bash
#!/bin/bash
export TF_CLOUD_ORG_NAME=sample-org
export TF_CLOUD_WS_NAME=sample-workspace
```

Then the command can be shorten to:

`
tfc-help update --var some_variable=some_value
`

Multiple variables can be passed to the command:

`
tfc-help update --var variable1=some_value,variable2=some_other_value
`

2. If you don't know whether the value has been created or not and you don't want to overwrite the existing value of the variable:

`
tfc-help update --var some_variable=some_value -k
`

3. Create/Update variable with non-default flags. Variable has value masked

`
tfc-help update --var some_variable=some_value -s
`

4. If an existing variable were configured to be in one category (For example, configured as terraform variable) but you want to change that to environment variable. Another use case would when a variable was marked as sensitive, but you want to change that to be non-sensitive. Use the -r flag to recreate the value and -k to keep the existing value. If -k flag is not used, then you have to specify the new value

Keep the existing value:

`
tfc-helper update --var test21 -r -k
`

Create new variable with the same name but new value

`
tfc-helper update --var test21=new_value -r
`

5. The tool can grab the existing environment variables starting with `TF_VAR_` in your machine and put them to Terraform Cloud.

For example, assuming that you have:

```bash
TF_VAR_variable1=something
TF_VAR_variable2=something_else
```

The following command can help put those variables to the cloud:

`
tfc-helper update -e
`

You can input both variables in the environment variables with the variables you put in the cli with the following command. Flag -r is to make sure all variables will replaced in case we need to update the categories of the variable as mentioned in Example Command #4

`
tfc-helper update -e --var "variable3=value" -r
`

If you want to put the variables as Terraform variables:

`
tfc-helper update -e --var "variable3=value" -t -r
`

6. Delete a single variable

`
tfc-helper delete --var some_variable
`

7. Delete all variables (Terraform and Environment variables) in Terraform Cloud

`
tfc-helper delete -a
`

## TODO:

- Configure CI
- Publish package
- Develop test cases
- Get environment variables for Windows
- Keep description of the variable when -k is not used
- The ability to take in a file with all variables needed and put to Terraform Cloud