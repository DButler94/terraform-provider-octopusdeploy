package octopusdeploy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/MattHodge/go-octopusdeploy/octopusdeploy"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"deployment_process_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"project_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_failure_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "EnvironmentDefault",
				ValidateFunc: validateValueFunc([]string{
					"EnvironmentDefault",
					"Off",
					"On",
				}),
			},
			"skip_machine_behavior": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "None",
				ValidateFunc: validateValueFunc([]string{
					"SkipUnavailableMachines",
					"None",
				}),
			},
			"deployment_step_windows_service": getDeploymentStepWindowsServiceSchema(),
			"deployment_step_iis_website":     getDeploymentStepIISWebsiteSchema(),
		},
	}
}

// addStandardDeploymentStepSchema adds the common schema for Octopus Deploy Steps
func addStandardDeploymentStepSchema(schemaToAddToo interface{}) *schema.Resource {
	schemaResource := schemaToAddToo.(*schema.Resource)

	schemaResource.Schema["configuration_transforms"] = &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	}

	schemaResource.Schema["configuration_variables"] = &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	}

	schemaResource.Schema["feed_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "feeds-builtin",
	}

	schemaResource.Schema["step_condition"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateFunc: validateValueFunc([]string{
			"success",
			"failure",
			"always",
			"variable",
		}),
		Default: "success",
	}

	schemaResource.Schema["step_name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}

	schemaResource.Schema["step_start_trigger"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "StartAfterPrevious",
		ValidateFunc: validateValueFunc([]string{
			"startafterprevious",
			"startwithprevious",
		}),
	}

	schemaResource.Schema["target_roles"] = &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	schemaResource.Schema["json_file_variable_replacement"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A comma-separated list of file names to replace settings in, relative to the package contents.",
	}

	return schemaResource
}

func addIISApplicationPoolSchema(schemaToAddToo interface{}) *schema.Resource {
	schemaResource := schemaToAddToo.(*schema.Resource)

	schemaResource.Schema["application_pool_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "Name of the application pool in IIS to create or reconfigure.",
		Required:    true,
	}

	schemaResource.Schema["application_pool_framework"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The version of the .NET common language runtime that this application pool will use. Choose v2.0 for applications built against .NET 2.0, 3.0 or 3.5. Choose v4.0 for .NET 4.0 or 4.5.",
		Default:     "v4.0",
		Optional:    true,
		ValidateFunc: validateValueFunc([]string{
			"v2.0",
			"v4.0",
		}),
	}

	schemaResource.Schema["application_pool_identity"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Which built-in account will the application pool run under.",
		Default:     "ApplicationPoolIdentity",
		ValidateFunc: validateValueFunc([]string{
			"ApplicationPoolIdentity",
			"LocalService",
			"LocalSystem",
			"NetworkService",
			"SpecificUser",
		}),
	}

	return schemaResource
}

// getDeploymentStepIISWebsiteSchema returns schema for an IIS deployment step
func getDeploymentStepIISWebsiteSchema() *schema.Schema {
	schmeaToReturn := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"anonymous_authentication": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether IIS should allow anonymous authentication.",
					Default:     false,
				},
				"basic_authentication": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether IIS should allow basic authentication with a 401 challenge.",
					Default:     false,
				},
				"website_name": {
					Type:        schema.TypeString,
					Description: "Create or update an IIS Web Site",
					Required:    true,
				},
				"windows_authentication": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether IIS should allow integrated Windows authentication with a 401 challenge.",
					Default:     true,
				},
			},
		},
	}

	schmeaToReturn.Elem = addStandardDeploymentStepSchema(schmeaToReturn.Elem)
	schmeaToReturn.Elem = addIISApplicationPoolSchema(schmeaToReturn.Elem)

	return schmeaToReturn
}

// getDeploymentStepWindowsServiceSchema returns schema for a Windows Service deployment step
func getDeploymentStepWindowsServiceSchema() *schema.Schema {
	schmeaToReturn := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"executable_path": {
					Type:     schema.TypeString,
					Required: true,
				},
				"service_account": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "LocalSystem",
				},
				"service_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"service_start_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "auto",
					ValidateFunc: validateValueFunc([]string{
						"auto",
						"delayed-auto",
						"demand",
						"unchanged",
					}),
				},
			},
		},
	}

	schmeaToReturn.Elem = addStandardDeploymentStepSchema(schmeaToReturn.Elem)

	return schmeaToReturn
}

func buildDeploymentProcess(d *schema.ResourceData, deploymentProcess *octopusdeploy.DeploymentProcess) *octopusdeploy.DeploymentProcess {
	deploymentProcess.Steps = nil // empty the steps

	if v, ok := d.GetOk("deployment_step_windows_service"); ok {
		steps := v.([]interface{})
		for _, raw := range steps {

			localStep := raw.(map[string]interface{})

			configurationTransforms := localStep["configuration_transforms"].(bool)
			configurationVariables := localStep["configuration_variables"].(bool)
			executablePath := localStep["executable_path"].(string)
			feedID := localStep["feed_id"].(string)
			jsonFileVariableReplacement := localStep["json_file_variable_replacement"].(string)
			serviceAccount := localStep["service_account"].(string)
			serviceName := localStep["service_name"].(string)
			serviceStartMode := localStep["service_start_mode"].(string)
			stepCondition := localStep["step_condition"].(string)
			stepName := localStep["step_name"].(string)
			stepStartTrigger := localStep["step_start_trigger"].(string)

			deploymentStep := &octopusdeploy.DeploymentStep{
				Name:               stepName,
				PackageRequirement: "LetOctopusDecide",
				Condition:          stepCondition,
				StartTrigger:       stepStartTrigger,
				Actions: []octopusdeploy.DeploymentAction{
					{
						Name:       stepName,
						ActionType: "Octopus.WindowsService",
						Properties: map[string]string{
							"Octopus.Action.WindowsService.CreateOrUpdateService":                       "True",
							"Octopus.Action.WindowsService.ServiceAccount":                              serviceAccount,
							"Octopus.Action.WindowsService.StartMode":                                   serviceStartMode,
							"Octopus.Action.Package.AutomaticallyRunConfigurationTransformationFiles":   strconv.FormatBool(configurationTransforms),
							"Octopus.Action.Package.AutomaticallyUpdateAppSettingsAndConnectionStrings": strconv.FormatBool(configurationVariables),
							"Octopus.Action.EnabledFeatures":                                            "Octopus.Features.WindowsService,Octopus.Features.ConfigurationTransforms,Octopus.Features.ConfigurationVariables",
							"Octopus.Action.Package.FeedId":                                             feedID,
							"Octopus.Action.Package.PackageId":                                          "a", // to fix
							"Octopus.Action.Package.DownloadOnTentacle":                                 "False",
							"Octopus.Action.WindowsService.ServiceName":                                 serviceName,
							"Octopus.Action.WindowsService.ExecutablePath":                              executablePath,
						},
					},
				},
			}

			if jsonFileVariableReplacement != "" {
				deploymentStep.Actions[0].Properties["Octopus.Action.Package.JsonConfigurationVariablesTargets"] = jsonFileVariableReplacement
				deploymentStep.Actions[0].Properties["Octopus.Action.Package.JsonConfigurationVariablesEnabled"] = "True"

				deploymentStep.Actions[0].Properties["Octopus.Action.EnabledFeatures"] += ",Octopus.Features.JsonConfigurationVariables"
			}

			if targetRolesInterface, ok := localStep["target_roles"]; ok {
				var targetRoleSlice []string

				targetRoles := targetRolesInterface.([]interface{})

				for _, role := range targetRoles {
					targetRoleSlice = append(targetRoleSlice, role.(string))
				}

				deploymentStep.Properties = map[string]string{"Octopus.Action.TargetRoles": strings.Join(targetRoleSlice, ",")}
			}

			deploymentProcess.Steps = append(deploymentProcess.Steps, *deploymentStep)
		}
	}

	if v, ok := d.GetOk("deployment_step_iis_website"); ok {
		steps := v.([]interface{})
		for _, raw := range steps {

			localStep := raw.(map[string]interface{})

			anonymousAuthentication := localStep["anonymous_authentication"].(bool)
			applicationPoolFramework := localStep["application_pool_framework"].(string)
			applicationPoolIdentity := localStep["application_pool_identity"].(string)
			applicationPoolName := localStep["application_pool_name"].(string)
			basicAuthentication := localStep["basic_authentication"].(bool)
			configurationTransforms := localStep["configuration_transforms"].(bool)
			configurationVariables := localStep["configuration_variables"].(bool)
			feedID := localStep["feed_id"].(string)
			jsonFileVariableReplacement := localStep["json_file_variable_replacement"].(string)
			stepCondition := localStep["step_condition"].(string)
			stepName := localStep["step_name"].(string)
			stepStartTrigger := localStep["step_start_trigger"].(string)
			websiteName := localStep["website_name"].(string)
			windowsAuthentication := localStep["windows_authentication"].(bool)

			deploymentStep := &octopusdeploy.DeploymentStep{
				Name:               stepName,
				PackageRequirement: "LetOctopusDecide",
				Condition:          stepCondition,
				StartTrigger:       stepStartTrigger,
				Actions: []octopusdeploy.DeploymentAction{
					{
						Name:       stepName,
						ActionType: "Octopus.IIS",
						Properties: map[string]string{
							"Octopus.Action.IISWebSite.DeploymentType":                                  "webSite",
							"Octopus.Action.IISWebSite.CreateOrUpdateWebSite":                           "True",
							"Octopus.Action.IISWebSite.Bindings":                                        "[{\"protocol\":\"http\",\"port\":\"80\",\"host\":\"\",\"thumbprint\":null,\"certificateVariable\":null,\"requireSni\":false,\"enabled\":true}]",
							"Octopus.Action.IISWebSite.ApplicationPoolFrameworkVersion":                 applicationPoolFramework,
							"Octopus.Action.IISWebSite.ApplicationPoolIdentityType":                     applicationPoolIdentity,
							"Octopus.Action.IISWebSite.EnableAnonymousAuthentication":                   strconv.FormatBool(anonymousAuthentication),
							"Octopus.Action.IISWebSite.EnableBasicAuthentication":                       strconv.FormatBool(basicAuthentication),
							"Octopus.Action.IISWebSite.EnableWindowsAuthentication":                     strconv.FormatBool(windowsAuthentication),
							"Octopus.Action.IISWebSite.WebApplication.ApplicationPoolFrameworkVersion":  applicationPoolFramework,
							"Octopus.Action.IISWebSite.WebApplication.ApplicationPoolIdentityType":      applicationPoolIdentity,
							"Octopus.Action.Package.AutomaticallyRunConfigurationTransformationFiles":   strconv.FormatBool(configurationTransforms),
							"Octopus.Action.Package.AutomaticallyUpdateAppSettingsAndConnectionStrings": strconv.FormatBool(configurationVariables),
							"Octopus.Action.EnabledFeatures":                                            "Octopus.Features.IISWebSite,Octopus.Features.ConfigurationTransforms,Octopus.Features.ConfigurationVariables",
							"Octopus.Action.Package.FeedId":                                             feedID,
							"Octopus.Action.Package.DownloadOnTentacle":                                 "False",
							"Octopus.Action.IISWebSite.WebRootType":                                     "packageRoot",
							"Octopus.Action.IISWebSite.StartApplicationPool":                            "True",
							"Octopus.Action.IISWebSite.StartWebSite":                                    "True",
							"Octopus.Action.Package.PackageId":                                          "a", // to fix
							"Octopus.Action.IISWebSite.WebSiteName":                                     websiteName,
							"Octopus.Action.IISWebSite.ApplicationPoolName":                             applicationPoolName,
						},
					},
				},
			}

			if jsonFileVariableReplacement != "" {
				deploymentStep.Actions[0].Properties["Octopus.Action.Package.JsonConfigurationVariablesTargets"] = jsonFileVariableReplacement
				deploymentStep.Actions[0].Properties["Octopus.Action.Package.JsonConfigurationVariablesEnabled"] = "True"

				deploymentStep.Actions[0].Properties["Octopus.Action.EnabledFeatures"] += ",Octopus.Features.JsonConfigurationVariables"
			}

			if targetRolesInterface, ok := localStep["target_roles"]; ok {
				var targetRoleSlice []string

				targetRoles := targetRolesInterface.([]interface{})

				for _, role := range targetRoles {
					targetRoleSlice = append(targetRoleSlice, role.(string))
				}

				deploymentStep.Properties = map[string]string{"Octopus.Action.TargetRoles": strings.Join(targetRoleSlice, ",")}
			}

			deploymentProcess.Steps = append(deploymentProcess.Steps, *deploymentStep)
		}
	}

	return deploymentProcess
}

func buildProjectResource(d *schema.ResourceData) *octopusdeploy.Project {
	name := d.Get("name").(string)
	lifecycleID := d.Get("lifecycle_id").(string)
	projectGroupID := d.Get("project_group_id").(string)

	project := octopusdeploy.NewProject(name, lifecycleID, projectGroupID)

	if attr, ok := d.GetOk("description"); ok {
		project.Description = attr.(string)
	}

	if attr, ok := d.GetOk("default_failure_mode"); ok {
		project.DefaultGuidedFailureMode = attr.(string)
	}

	if attr, ok := d.GetOk("skip_machine_behavior"); ok {
		project.ProjectConnectivityPolicy.SkipMachineBehavior = attr.(string)
	}

	return project
}

func updateDeploymentProcess(d *schema.ResourceData, client *octopusdeploy.Client, projectID string) error {
	deploymentProcess, err := client.DeploymentProcess.Get(projectID)

	if err != nil {
		return fmt.Errorf("error getting deployment process for project: %s", err.Error())
	}

	newDeploymentProcess := buildDeploymentProcess(d, deploymentProcess)
	// set the newly build deployment processes ID so it can be updated
	newDeploymentProcess.ID = deploymentProcess.ID

	updateDeploymentProcess, err := client.DeploymentProcess.Update(newDeploymentProcess)

	if err != nil {
		return fmt.Errorf("error creating deployment process for project: %s", err.Error())
	}

	d.Set("deployment_process_id", updateDeploymentProcess.ID)

	return nil
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*octopusdeploy.Client)

	newProject := buildProjectResource(d)

	createdProject, err := client.Project.Add(newProject)

	if err != nil {
		return fmt.Errorf("error creating project: %s", err.Error())
	}

	d.SetId(createdProject.ID)

	// set the deployment process
	errUpdatingDeploymentProcess := updateDeploymentProcess(d, client, createdProject.DeploymentProcessID)

	// deployment process is updated, not created, but log message makes more sense if it fails in a create step
	if errUpdatingDeploymentProcess != nil {
		return fmt.Errorf("error creating deploymentprocess: %s", errUpdatingDeploymentProcess.Error())
	}

	return nil
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*octopusdeploy.Client)

	projectID := d.Id()

	project, err := client.Project.Get(projectID)

	if err == octopusdeploy.ErrItemNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading project id %s: %s", projectID, err.Error())
	}

	log.Printf("[DEBUG] project: %v", m)
	d.Set("name", project.Name)
	d.Set("description", project.Description)
	d.Set("lifecycle_id", project.LifecycleID)
	d.Set("project_group_id", project.ProjectGroupID)
	d.Set("default_failure_mode", project.DefaultGuidedFailureMode)
	d.Set("skip_machine_behavior", project.ProjectConnectivityPolicy.SkipMachineBehavior)

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	project := buildProjectResource(d)
	project.ID = d.Id() // set project struct ID so octopus knows which project to update

	client := m.(*octopusdeploy.Client)

	project, err := client.Project.Update(project)

	if err != nil {
		return fmt.Errorf("error updating project id %s: %s", d.Id(), err.Error())
	}

	d.SetId(project.ID)

	// set the deployment process
	errUpdatingDeploymentProcess := updateDeploymentProcess(d, client, project.DeploymentProcessID)

	// deployment process is updated, not created, but log message makes more sense if it fails in a create step
	if errUpdatingDeploymentProcess != nil {
		return fmt.Errorf("error creating deploymentprocess: %s", errUpdatingDeploymentProcess.Error())
	}

	return nil
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*octopusdeploy.Client)

	projectID := d.Id()

	err := client.Project.Delete(projectID)

	if err != nil {
		return fmt.Errorf("error deleting project id %s: %s", projectID, err.Error())
	}

	d.SetId("")
	return nil
}
