package nodes

import (
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/aggregate"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/code"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/delete_execution"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/filter"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/html"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/http_request"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/if"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/limit"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/manual_trigger"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/respond_to_webhook"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/schedule_trigger"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/switch"
	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/webhook"
)
