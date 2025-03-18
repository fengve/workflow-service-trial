package api

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func (service *WorkflowService) GetDynamicNodeParameters_Options(c *fiber.Ctx) error {
	orgId := c.Params("orgId")
	sugerOrgId := c.Query("sugerOrgId")
	if orgId == "" || sugerOrgId == "" {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId or sugerOrgId is empty"))
	}
	if orgId != sugerOrgId {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId and sugerOrgId are not matched"))
	}

	request := &structs.GetWorkflowDynamicNodeParametersRequest{}
	if err := c.QueryParser(request); err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	request.CurrentNodeParameters = parseCurrentNodeParametersFromQuerys(c.Queries())

	response, err := getOptions(c.UserContext(), request)

	if err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) GetDynamicNodeParameters_ResourceLocatorResults(c *fiber.Ctx) error {
	orgId := c.Params("orgId")
	sugerOrgId := c.Query("sugerOrgId")
	if orgId == "" || sugerOrgId == "" {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId or sugerOrgId is empty"))
	}
	if orgId != sugerOrgId {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId and sugerOrgId are not matched"))
	}

	request := &structs.GetWorkflowDynamicNodeParametersRequest{}
	if err := c.QueryParser(request); err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	request.CurrentNodeParameters = parseCurrentNodeParametersFromQuerys(c.Queries())

	response, err := getResourceLocatorResults(c.UserContext(), request)
	if err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (service *WorkflowService) GetDynamicNodeParameters_ResourceMapperFields(c *fiber.Ctx) error {
	orgId := c.Params("orgId")
	sugerOrgId := c.Query("sugerOrgId")
	if orgId == "" || sugerOrgId == "" {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId or sugerOrgId is empty"))
	}
	if orgId != sugerOrgId {
		return HandleBadRequestErrorWithTrace(c, fmt.Errorf("orgId and sugerOrgId are not matched"))
	}

	request := &structs.GetWorkflowDynamicNodeParametersRequest{}
	if err := c.QueryParser(request); err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	request.CurrentNodeParameters = parseCurrentNodeParametersFromQuerys(c.Queries())

	response, err := getResourceMappingFields(c.UserContext(), request)

	if err != nil {
		return HandleInternalServerErrorWithTrace(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func getOptions(
	ctx context.Context,
	request *structs.GetWorkflowDynamicNodeParametersRequest) (*structs.GetDynamicNodeParametersResponse_Options, error) {
	result := &structs.GetDynamicNodeParametersResponse_Options{}
	nodeName := request.NodeTypeAndVersion.Name
	methodName := request.MethodName
	// Get node
	nodeMethods, err := findNodeMethods(nodeName)
	if err != nil {
		return nil, err
	}
	// Get method and call it
	loadOptionsMethods := nodeMethods.Methods().LoadOptions
	if method, ok := loadOptionsMethods[methodName]; ok {
		callResult, err := method(ctx, request.SugerOrgId, request.CurrentNodeParameters)
		if err != nil {
			return nil, fmt.Errorf("the method %s execute error: %s", methodName, err.Error())
		}
		result = callResult
		return result, nil
	} else {
		return nil, fmt.Errorf("the node %s does not have the method of %s", nodeName, methodName)
	}
}

func getResourceMappingFields(
	ctx context.Context,
	request *structs.GetWorkflowDynamicNodeParametersRequest,
) (*structs.GetDynamicNodeParametersResponse_ResourceMapperFields, error) {
	result := &structs.GetDynamicNodeParametersResponse_ResourceMapperFields{}
	nodeName := request.NodeTypeAndVersion.Name
	methodName := request.MethodName
	// Get node
	nodeMethods, err := findNodeMethods(nodeName)
	if err != nil {
		return nil, err
	}
	// Get method and call it
	resourceMappingMethods := nodeMethods.Methods().ResourceMapping
	if method, ok := resourceMappingMethods[methodName]; ok {
		callResult, err := method(ctx, request.SugerOrgId, request.CurrentNodeParameters)
		if err != nil {
			return nil, fmt.Errorf("the method %s execute error: %s", methodName, err.Error())
		}
		result = callResult
		return result, nil
	} else {
		return nil, fmt.Errorf("the node %s does not have the method of %s", nodeName, methodName)
	}
}

func getResourceLocatorResults(
	ctx context.Context,
	request *structs.GetWorkflowDynamicNodeParametersRequest,
) (*structs.GetDynamicNodeParametersResponse_ResourceLocatorResults, error) {
	result := &structs.GetDynamicNodeParametersResponse_ResourceLocatorResults{}
	nodeName := request.NodeTypeAndVersion.Name
	methodName := request.MethodName
	// Get node
	nodeMethods, err := findNodeMethods(nodeName)
	if err != nil {
		return nil, err
	}
	// Get method and call it
	listSearchMethods := nodeMethods.Methods().ListSearch
	if method, ok := listSearchMethods[methodName]; ok {
		callResult, err := method(ctx, request.SugerOrgId, request.CurrentNodeParameters, request.Filter, request.PaginationToken)
		if err != nil {
			return nil, fmt.Errorf("the method %s execute error: %s", methodName, err.Error())
		}
		result = callResult
		return result, nil
	} else {
		return nil, fmt.Errorf("the node %s does not have the method of %s", nodeName, methodName)
	}
}

// Find NodeObject by name and convert it to NodeMethods after check
func findNodeMethods(nodeName string) (core.NodeMethods, error) {
	nodeObject := core.MustNewNode(nodeName)
	if nodeObject == nil {
		return nil, fmt.Errorf("the node %s does not exist", nodeName)
	}

	var nodeMethods core.NodeMethods
	if newObject, ok := nodeObject.(core.NodeMethods); ok {
		nodeMethods = newObject
	} else {
		return nil, fmt.Errorf("the node %s does not have any methods", nodeName)
	}
	return nodeMethods, nil
}

// Parse currentNodeParameters from c.Queries() format to map[string]interface{}.
// E.g.
// input:
//
//	{
//		"currentNodeParameters[operation]":"executeQuery",
//		"currentNodeParameters[options][includeSchema]":"false",
//		"currentNodeParameters[projectId][__rl]":"true",
//		"currentNodeParameters[projectId][mode]":"list",
//		"currentNodeParameters[projectId][value]":"suger-dev",
//		"currentNodeParameters[resource]":"database"
//	}
//
// output:
//
//	{
//		"operation":"executeQuery",
//	 	"options": {
//			"includeSchema": "false"
//		},
//		"projectId": {
//			"__rl": "true",
//			"mode": "list",
//			"value":"suger-dev"
//		},
//		"resource": "database"
//	}
func parseCurrentNodeParametersFromQuerys(querys map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	pattern := `\[([^\]]+)\]`
	reg := regexp.MustCompile(pattern)
	for key, value := range querys {
		if !strings.HasPrefix(key, "currentNodeParameters") {
			continue
		}
		matches := reg.FindAllStringSubmatch(key, -1)
		current := &result
		len := len(matches)
		for index, match := range matches {
			name := match[1]
			if valExist, ok := (*current)[name]; ok {
				if index == len-1 {
					fmt.Println("bad case, the key already exists")
				} else {
					m, ok := valExist.(map[string]interface{})
					if !ok {
						fmt.Println("bad case, the key already exists but is not a map")
					}
					current = &m
				}
			} else {
				if index == len-1 {
					(*current)[name] = value
				} else {
					child := make(map[string]interface{})
					(*current)[name] = child
					current = &child
				}
			}
		}
	}
	return result
}

func (service *WorkflowService) RegisterRouteMethods_DynamicParameter() {
	service.fiberApp.Get("/workflow/org/:orgId/workflow/dynamic-node-parameters/options", service.GetDynamicNodeParameters_Options)
	service.fiberApp.Get("/workflow/org/:orgId/workflow/dynamic-node-parameters/resource-locator-results", service.GetDynamicNodeParameters_ResourceLocatorResults)
	service.fiberApp.Get("/workflow/org/:orgId/workflow/dynamic-node-parameters/resource-mapper-fields", service.GetDynamicNodeParameters_ResourceMapperFields)
}
