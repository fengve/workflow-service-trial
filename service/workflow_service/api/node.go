package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const N8N_ICON_PREFIX = "n8n-nodes-base/dist/nodes/"
const EMBED_ICON_PREFIX = "embed/"
const CONTENT_TYPE_SVG = "image/svg+xml"
const CONTENT_TYPE_PNG = "image/png"

// Save the all node json configs.
var nodeJsonConfigs []interface{}

func (service *WorkflowService) GetWorkflowNodesJson(ctx *fiber.Ctx) error {
	// If the nodeJsonConfigs is empty, we need to generate the json file.
	if len(nodeJsonConfigs) == 0 {
		nodesMap := core.GetAllNodeObjects()
		// It is safe to ignore all the errors, because we generate the json file in the building process.
		for _, node := range nodesMap {
			spec, ok := node.DefaultSpec().(*structs.WorkflowNodeSpec)
			if !ok {
				return HandleInternalServerErrorWithTrace(
					ctx, fmt.Errorf("Failed to convert to node json Spec: %s %s", node.Name(), node.Category()))
			}
			var nodeJsonConfig interface{}
			err := json.Unmarshal(spec.JsonConfig, &nodeJsonConfig)
			if err != nil {
				return HandleInternalServerErrorWithTrace(
					ctx, fmt.Errorf("Failed to unmarshal json config: %s %s with error: %v", node.Name(), node.Category(), err))
			}
			if nodeJsonConfig == nil {
				return HandleInternalServerErrorWithTrace(
					ctx, fmt.Errorf("Failed to unmarshal json config: %s %s", node.Name(), node.Category()))
			}
			nodeJsonConfigs = append(nodeJsonConfigs, nodeJsonConfig)
		}
	}

	ctx.Set("Content-Type", "application/json")
	return ctx.Status(fiber.StatusOK).JSON(nodeJsonConfigs)
}

func (service *WorkflowService) GetNodeIcons(ctx *fiber.Ctx) error {
	path := ctx.Params("+")

	// check path is remote or local embedding
	if strings.HasPrefix(path, N8N_ICON_PREFIX) {
		// n8n-nodes-base/dist/nodes/Suger/suger.svg
		return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Errorf("n8n node icon not supported:%s", path).Error())
	}
	if strings.HasPrefix(path, EMBED_ICON_PREFIX) {
		// local embedding
		// embed/n8n-nodes-base.code/code.svg
		nodeName, nodeIconType, ok := getNodeNameAndIconType(path)
		if !ok {
			return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Errorf("invalid path:%s", path).Error())
		}
		nodeEmbedIcons := core.GetAllNodeEmbedIcons()
		nodeIcon, ok := nodeEmbedIcons[nodeName]
		if !ok {
			// not found
			return ctx.Status(fiber.StatusNotFound).SendString(fmt.Errorf("icon not found:%s", path).Error())
		}
		contentType := getContentType(nodeIconType)
		ctx.Set("Content-Type", contentType)
		_, err := ctx.Write(nodeIcon)
		return err
	}
	// invalid path
	return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Errorf("invalid path:%s", path).Error())
}

func getContentType(iconType string) string {
	if iconType == "svg" {
		return CONTENT_TYPE_SVG
	}
	if iconType == "png" {
		return CONTENT_TYPE_PNG
	}
	return ""
}

func getNodeNameAndIconType(path string) (nodeName, iconType string, ok bool) {
	// embed/n8n-nodes-base.code/code.svg
	items := strings.Split(path, "/")
	if len(items) != 3 {
		return
	}
	nodeName = items[1]
	icon := items[2]
	iconItems := strings.Split(icon, ".")
	if len(iconItems) != 2 {
		return
	}

	iconType = iconItems[1]

	if iconType != "svg" && iconType != "png" {
		return
	}
	ok = true
	return
}

func (service *WorkflowService) RegisterRouteMethods_Node() {
	service.fiberApp.Get("/workflow/public/nodes.json", service.GetWorkflowNodesJson)
	service.fiberApp.Get("/workflow/public/nodes/icons/+", service.GetNodeIcons)
}
