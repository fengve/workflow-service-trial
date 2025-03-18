package nodes_test

// Command to run this test only
// go test -v service/workflow_service/nodes_test/init_test.go  service/workflow_service/nodes_test/schedule_trigger_test.go

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/api"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/schedule_trigger"
	"github.com/sugerio/workflow-service-trial/shared"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	sharedTemporal "github.com/sugerio/workflow-service-trial/shared/temporal"
)

type ScheduleTriggerTestSuite struct {
	suite.Suite
}

func Test_ScheduleTrigger(t *testing.T) {
	suite.Run(t, new(ScheduleTriggerTestSuite))
}

func (s *ScheduleTriggerTestSuite) Test() {
	s.T().Run("TestScheduleTriggerSpec", func(t *testing.T) {
		assert := require.New(s.T())

		var scheduleTriggerSpec structs.WorkflowNodeDescriptionSpec
		testFile, err := os.ReadFile("./test_files/schedule-spec.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &scheduleTriggerSpec)
		assert.Nil(err)

		assert.Empty(scheduleTriggerSpec.Inputs)
		assert.NotEmpty(scheduleTriggerSpec.ActivationMessage)
		assert.Equal(float64(1), scheduleTriggerSpec.Version.([]interface{})[0])
		value, ok := scheduleTriggerSpec.Properties[1].Options[0].(map[string]interface{})
		assert.True(ok)
		assert.Equal("interval", value["name"])
		values, ok := value["values"].([]interface{})
		assert.True(ok)
		assert.Equal("field", values[0].(map[string]interface{})["name"])
	})

	s.T().Run("TestScheduleTriggerGenerate", func(t *testing.T) {
		assert := require.New(s.T())

		_ = schedule_trigger.ScheduleTrigger{}
		executor := core.NewExecutor(schedule_trigger.Name)
		node := executor.GetNode()
		assert.Equal("Schedule Trigger", node.DefaultSpec().(*structs.WorkflowNodeSpec).NodeSpec.DisplayName)

		var scheduleParams schedule_trigger.ScheduleParams
		testFile, err := os.ReadFile("./test_files/schedule-params.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &scheduleParams)
		assert.Nil(err)
		assert.Equal(7, len(scheduleParams.Rule.Interval))
		assert.Equal([]int{0, 1}, scheduleParams.Rule.Interval[0].TriggerAtDay)
		assert.Equal("15 * * * *", scheduleParams.Rule.Interval[1].Expression)
		assert.Equal(1, scheduleParams.Rule.Interval[2].TriggerAtHour)
		assert.Equal(2, scheduleParams.Rule.Interval[3].TriggerAtDayOfMonth)
		assert.Equal("hours", scheduleParams.Rule.Interval[4].Field)
		assert.Equal(15, scheduleParams.Rule.Interval[5].MinutesInterval)
		assert.Equal(35, scheduleParams.Rule.Interval[6].SecondsInterval)

		// Test set default
		var params map[string]interface{}
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		st := node.(*schedule_trigger.ScheduleTrigger)
		intervals := st.SetDefaultValues(params["rule"].(map[string]interface{})["interval"].([]interface{}))
		assert.Equal(7, len(intervals))
		assert.Equal("days", intervals[2].Field)

		// Test set default emtpy
		testFile, err = os.ReadFile("./test_files/schedule-params_empty.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		intervals = st.SetDefaultValues(params["rule"].(map[string]interface{})["interval"].([]interface{}))
		assert.Equal(1, len(intervals))
		assert.Equal("days", intervals[0].Field)

		// Test set default weeks
		testFile, err = os.ReadFile("./test_files/schedule-params_weeks.json")
		assert.Nil(err)
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		intervals = st.SetDefaultValues(params["rule"].(map[string]interface{})["interval"].([]interface{}))
		assert.Equal(1, len(intervals))
		assert.Equal(0, intervals[0].TriggerAtHour)
		assert.Equal([]int{}, intervals[0].TriggerAtDay)
	})

	s.T().Run("TestScheduleTriggerTrigger", func(t *testing.T) {
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}
		assert := require.New(s.T())

		manualRunReq := &structs.WorkflowManualRunRequest{}
		testFileList := []string{
			"./test_files/schedule-real-run.json",
		}
		ctx := context.Background()

		for _, file := range testFileList {
			testFile, err := os.ReadFile(file)
			assert.Nil(err)

			err = json.Unmarshal(testFile, manualRunReq)
			assert.Nil(err)

			_, err = rdsDbQueries.CreateWorkflowEntity(ctx, lib.CreateWorkflowEntityParams{
				Name:        manualRunReq.WorkflowData.Name,
				Active:      manualRunReq.WorkflowData.Active,
				Nodes:       json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Nodes)),
				Connections: json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Connections)),
				Settings: pqtype.NullRawMessage{
					RawMessage: json.RawMessage(core.JsonStr(manualRunReq.WorkflowData.Settings)),
					Valid:      true},
				StaticData: pqtype.NullRawMessage{Valid: false},
				PinData: pqtype.NullRawMessage{
					RawMessage: json.RawMessage("{}"),
					Valid:      true},
				VersionId:    sql.NullString{String: manualRunReq.WorkflowData.VersionId, Valid: true}, // new versionid.
				TriggerCount: 0,
				ID:           manualRunReq.WorkflowData.ID,
				Meta:         pqtype.NullRawMessage{Valid: false},
				SugerOrgId:   "",
			})
			assert.Nil(err)
		}
	})

	s.T().Run("TestRecordGenerateCronExpression", func(t *testing.T) {
		assert := require.New(s.T())

		_ = schedule_trigger.ScheduleTrigger{}
		executor := core.NewExecutor(schedule_trigger.Name)
		node := executor.GetNode()
		st := node.(*schedule_trigger.ScheduleTrigger)

		testFile, err := os.ReadFile("./test_files/schedule-params.json")
		assert.Nil(err)
		var params map[string]interface{}
		err = json.Unmarshal(testFile, &params)
		assert.Nil(err)
		intervals := st.SetDefaultValues(params["rule"].(map[string]interface{})["interval"].([]interface{}))
		assert.Equal(7, len(intervals))
		expression := st.GenerateCronExpression(intervals[0])
		assert.Equal("15 1 * * 0,1", expression)
		expression = st.GenerateCronExpression(intervals[1])
		assert.Equal("15 * * * *", expression)
		expression = st.GenerateCronExpression(intervals[2])
		assert.Equal("0 1 */2 * *", expression)
		expression = st.GenerateCronExpression(intervals[3])
		assert.Equal("10 1 2 */2 *", expression)
		expression = st.GenerateCronExpression(intervals[4])
		assert.Equal("10 */2 * * *", expression)
		expression = st.GenerateCronExpression(intervals[5])
		assert.Equal("@every 15m", expression)
		expression = st.GenerateCronExpression(intervals[6])
		assert.Equal("@every 35s", expression)
	})

	s.T().Run(("TestScheduleTrigger Workflow Activate Deactivate"), func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}

		assert := require.New(s.T())
		ctx := context.Background()
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/schedule-trigger-workflow.json")
		assert.Nil(err)
		workflowID := newWorkflow.ID

		// Activate the workflow
		err = api.ActivateWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
		// Wait for 2 seconds to let the cron temporal workflow scheduled.
		time.Sleep(2 * time.Second)

		// Check the temporal workflow scheduled
		temporalWorkflowExecutions, err := sharedTemporal.ListOpenWorkflowExecutionsByOrgAndType(
			ctx, temporalClient, organization.ID, "Workflow_ScheduleTrigger")
		assert.Nil(err)
		assert.Len(temporalWorkflowExecutions, 1)

		// Wait for 4 seconds to let the cron temporal workflow run.
		time.Sleep(4 * time.Second)

		// Deactivate the workflow
		err = api.DeactivateWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)

		// Check Execution Entity
		count, err := rdsDbQueries.CountWorkflowExecutionEntitiesByWorkflowId(context.Background(), workflowID)
		assert.Nil(err)
		// Check DeleteExecution hook. The count shall be zero since the execution is deleted.
		assert.Equal(count, int64(0))

		// Check the scheduled temporal workflow has been terminated.
		temporalWorkflowExecutions, err = sharedTemporal.ListOpenWorkflowExecutionsByOrgAndType(
			ctx, temporalClient, organization.ID, "Workflow_ScheduleTrigger")
		assert.Nil(err)
		assert.Len(temporalWorkflowExecutions, 0)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, workflowID)
		assert.Nil(err)
	})

	s.T().Run(("TestScheduleTrigger Workflow ManualRun"), func(t *testing.T) {
		t.Parallel()
		if environment.Env != shared.ENV_LOCAL_TEST {
			t.Skip()
		}

		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")
		// Create Workflow
		newWorkflow, err := api.CreateWorkflow_Testing(
			testFiberLambda, organization.ID, "./test_files/cron-dev-case-e2e.json")
		assert.Nil(err)

		// Manual Run the workflow
		resp, err := api.ManualRunWorkflow_Testing(testFiberLambda, newWorkflow)
		assert.Nil(err)
		assert.NotEmpty(resp)

		// Delete workflow
		err = api.DeleteWorkflow_Testing(testFiberLambda, organization.ID, newWorkflow.ID)
		assert.Nil(err)
	})
}
