package octopusdeploy

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/MattHodge/go-octopusdeploy/octopusdeploy"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccOctopusDeployProjectBasic(t *testing.T) {
	const terraformNamePrefix = "octopusdeploy_project.foo"
	const projectName = "Funky Monkey"
	const lifeCycleID = "Lifecycles-1"
	const projectGroupID = "ProjectGroups-1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOctopusDeployProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectBasic(projectName, lifeCycleID, projectGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOctopusDeployProjectExists(terraformNamePrefix),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "name", projectName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "lifecycle_id", lifeCycleID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "project_group_id", projectGroupID),
				),
			},
		},
	})
}

func TestAccOctopusDeployProjectWithDeploymentStepWindowsService(t *testing.T) {
	const terraformNamePrefix = "octopusdeploy_project.foo"
	const projectName = "Funky Monkey"
	const lifeCycleID = "Lifecycles-1"
	const projectGroupID = "ProjectGroups-1"
	const serviceName = "Epic Service"
	const executablePath = `bin\\MyService.exe` // needs 4 slashes to appear in the TF config as a double slash
	const stepName = "Deploying Epic Service"
	targetRoles := []string{"Lab1", "Lab2"}
	projectIDRegex, _ := regexp.Compile("Projects\\-")
	deploymentProcessIDRegex, _ := regexp.Compile("deploymentprocess\\-Projects\\-.*")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOctopusDeployProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithDeploymentStepWindowsService(projectName, lifeCycleID, projectGroupID, serviceName, executablePath, stepName, targetRoles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOctopusDeployProjectExists(terraformNamePrefix),
					resource.TestMatchResourceAttr(
						terraformNamePrefix, "id", projectIDRegex),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "name", projectName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "lifecycle_id", lifeCycleID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "project_group_id", projectGroupID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.service_name", serviceName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.step_name", stepName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.target_roles.0", targetRoles[0]),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.target_roles.1", targetRoles[1]),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.executable_path", strings.Replace(executablePath, "\\\\", "\\", 1)), // need to scape the backslashes
					resource.TestMatchResourceAttr(
						terraformNamePrefix, "deployment_process_id", deploymentProcessIDRegex),
				),
			},
		},
	})
}

func TestAccOctopusDeployProjectWithUpdate(t *testing.T) {
	const terraformNamePrefix = "octopusdeploy_project.foo"
	const projectName = "Funky Monkey"
	const lifeCycleID = "Lifecycles-1"
	const projectGroupID = "ProjectGroups-1"
	const description = "I am a new description"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOctopusDeployProjectDestroy,
		Steps: []resource.TestStep{
			// create project with no description
			{
				Config: testAccProjectBasic(projectName, lifeCycleID, projectGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOctopusDeployProjectExists(terraformNamePrefix),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "name", projectName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "lifecycle_id", lifeCycleID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "project_group_id", projectGroupID),
				),
			},
			// create update it with a description + build steps
			{
				Config: testAccWithMultipleDeploymentStepWindowsService,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOctopusDeployProjectExists(terraformNamePrefix),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "name", "Project Name"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "lifecycle_id", "Lifecycles-1"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "project_group_id", "ProjectGroups-1"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "description", "My Awesome Description"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.service_name", "My First Service"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.step_name", "Deploy My First Service"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.target_roles.0", "Role1"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.target_roles.1", "Role2"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.executable_path", "C:\\MyService\\my_service.exe"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.service_name", "My Second Service"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.step_name", "Deploy My Second Service"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.target_roles.0", "Role3"),
					resource.TestCheckNoResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.target_roles.1"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.executable_path", "C:\\MyService\\my_service2.exe"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.configuration_transforms", "false"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.configuration_variables", "false"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.step_name", "Deploy Website"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.target_roles.0", "MyRole1"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.website_name", "Awesome Website"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.application_pool_name", "MyAppPool"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.application_pool_framework", "v2.0"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.step_condition", "failure"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.basic_authentication", "true"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.anonymous_authentication", "true"),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.json_file_variable_replacement", "appsettings.json,Config\\*.json"),
				),
			},
			// update again by remove its description
			{
				Config: testAccProjectBasic(projectName, lifeCycleID, projectGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOctopusDeployProjectExists(terraformNamePrefix),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "name", projectName),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "lifecycle_id", lifeCycleID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "project_group_id", projectGroupID),
					resource.TestCheckResourceAttr(
						terraformNamePrefix, "description", ""),
					resource.TestCheckNoResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.0.step_name"),
					resource.TestCheckNoResourceAttr(
						terraformNamePrefix, "deployment_step_windows_service.1.step_name"),
					resource.TestCheckNoResourceAttr(
						terraformNamePrefix, "deployment_step_iis_website.0.step_name"),
				),
			},
		},
	})
}

func testAccProjectBasic(name, lifeCycleID, projectGroupID string) string {
	return fmt.Sprintf(`
		resource "octopusdeploy_project" "foo" {
			name           = "%s"
			lifecycle_id    = "%s"
			project_group_id = "%s"
		}
		`,
		name, lifeCycleID, projectGroupID,
	)
}

const testAccWithMultipleDeploymentStepWindowsService = `
resource "octopusdeploy_project" "foo" {
	name             = "Project Name"
	lifecycle_id     = "Lifecycles-1"
	project_group_id = "ProjectGroups-1"
	description      = "My Awesome Description"

	deployment_step_windows_service {
		executable_path          = "C:\\MyService\\my_service.exe"
		service_name             = "My First Service"
		step_name                = "Deploy My First Service"

		target_roles = [
		  "Role1",
		  "Role2"
		]
	}

	deployment_step_windows_service {
		executable_path          = "C:\\MyService\\my_service2.exe"
		service_name             = "My Second Service"
		step_name                = "Deploy My Second Service"
		step_start_trigger       = "startwithprevious"
		service_account          = "NewServiceAccount"
		service_start_mode       = "demand"
		configuration_transforms = false
		configuration_variables  = false


		target_roles = [
		  "Role3",
		]
	}

	deployment_step_iis_website {
		step_name                      = "Deploy Website"
		website_name                   = "Awesome Website"
		application_pool_name          = "MyAppPool"
		application_pool_framework     = "v2.0"
		step_condition                 = "failure"
		basic_authentication           = true
		anonymous_authentication       = true
		json_file_variable_replacement = "appsettings.json,Config\\*.json"

		target_roles = [
		  "MyRole1",
		]
	}
}
`

func testAccWithDeploymentStepWindowsService(name, lifeCycleID, projectGroupID, serviceName, executablePath, stepName string, targetRoles []string) string {
	return fmt.Sprintf(`
		resource "octopusdeploy_project" "foo" {
			name           = "%s"
			lifecycle_id    = "%s"
			project_group_id = "%s"

			deployment_step_windows_service {
				executable_path          = "%s"
				service_name             = "%s"
				step_name                = "%s"

				target_roles = [
				  "%s",
				]
			}
		}
		`,
		name, lifeCycleID, projectGroupID, executablePath, serviceName, stepName, strings.Join(targetRoles, "\",\""),
	)
}

func testAccCheckOctopusDeployProjectDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*octopusdeploy.Client)

	if err := destroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckOctopusDeployProjectExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*octopusdeploy.Client)
		if err := existsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func destroyHelper(s *terraform.State, client *octopusdeploy.Client) error {
	for _, r := range s.RootModule().Resources {
		if _, err := client.Project.Get(r.Primary.ID); err != nil {
			if err == octopusdeploy.ErrItemNotFound {
				continue
			}
			return fmt.Errorf("Received an error retrieving project %s", err)
		}
		return fmt.Errorf("Project still exists")
	}
	return nil
}

func existsHelper(s *terraform.State, client *octopusdeploy.Client) error {
	for _, r := range s.RootModule().Resources {
		if _, err := client.Project.Get(r.Primary.ID); err != nil {
			return fmt.Errorf("Received an error retrieving project %s", err)
		}
	}
	return nil
}
