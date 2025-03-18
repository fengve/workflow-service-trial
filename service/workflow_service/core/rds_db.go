package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

func GetWorkflowEntityById(ctx context.Context, workflowId string) (*structs.WorkflowEntity, error) {
	workflowEntity_RdsDbLib, err := rdsDbQueries.GetWorkflowEntityById(ctx, workflowId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such workflow")
		}
		return nil, err
	}
	workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
	if err != nil {
		return nil, err
	}
	return &workflowEntity, nil
}

// Get the workflow entity by orgId and workflowId.
// If the workflow does not exist, return an error.
func GetWorkflowEntity(ctx context.Context, orgId string, workflowId string) (*structs.WorkflowEntity, error) {
	workflowEntity_RdsDbLib, err := rdsDbQueries.GetWorkflowEntity(
		ctx,
		rdsDbLib.GetWorkflowEntityParams{
			SugerOrgId: orgId,
			ID:         workflowId,
		})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such workflow")
		}
		return nil, err
	}
	workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
	if err != nil {
		return nil, err
	}
	return &workflowEntity, nil
}

// List all active workflow entities by orgId.
func ListActiveWorkflowEntities(ctx context.Context, orgId string) ([]structs.WorkflowEntity, error) {
	workflowEntities_RdsDbLib, err := rdsDbQueries.ListActiveWorkflowEntities(ctx, orgId)
	if err != nil {
		return nil, err
	}
	workflowEntities := make([]structs.WorkflowEntity, 0)
	for _, workflowEntity_RdsDbLib := range workflowEntities_RdsDbLib {
		workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
		if err != nil {
			return nil, err
		}
		workflowEntities = append(workflowEntities, workflowEntity)
	}
	return workflowEntities, nil
}

// List all active workflow entities, regardless of orgId.
func ListAllActiveWorkflowEntities(ctx context.Context) ([]structs.WorkflowEntity, error) {
	workflowEntities_RdsDbLib, err := rdsDbQueries.ListAllActiveWorkflowEntities(ctx)
	if err != nil {
		return nil, err
	}
	workflowEntities := make([]structs.WorkflowEntity, 0)
	for _, workflowEntity_RdsDbLib := range workflowEntities_RdsDbLib {
		workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
		if err != nil {
			return nil, err
		}
		workflowEntities = append(workflowEntities, workflowEntity)
	}
	return workflowEntities, nil
}

// List all workflow entities by orgId.
func ListWorkflowEntities(ctx context.Context, orgId string) ([]structs.WorkflowEntity, error) {
	workflowEntities_RdsDbLib, err := rdsDbQueries.ListWorkflowEntities(ctx, orgId)
	if err != nil {
		return nil, err
	}
	workflowEntities := make([]structs.WorkflowEntity, 0)
	for _, workflowEntity_RdsDbLib := range workflowEntities_RdsDbLib {
		workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
		if err != nil {
			return nil, err
		}
		workflowEntities = append(workflowEntities, workflowEntity)
	}
	return workflowEntities, nil
}

func DeleteWorkflowEntity(ctx context.Context, orgId string, workflowId string) (*structs.WorkflowEntity, error) {
	workflowEntity_RdsDbLib, err := rdsDbQueries.DeleteWorkflowEntity(ctx, rdsDbLib.DeleteWorkflowEntityParams{
		SugerOrgId: orgId,
		ID:         workflowId,
	})
	if err != nil {
		return nil, err
	}
	workflowEntity, err := structs.ToWorkflowEntity(workflowEntity_RdsDbLib)
	if err != nil {
		return nil, err
	}
	return &workflowEntity, nil
}

func GetWorkflowExecution(ctx context.Context, executionId int32) (*structs.WorkflowExecution, error) {
	executionEntity, err := rdsDbQueries.GetWorkflowExecutionEntity(ctx, executionId)
	if err != nil {
		return nil, err
	}
	executionData, err := rdsDbQueries.GetWorkflowExecutionData(ctx, executionId)
	if err != nil {
		return nil, err
	}
	return toWorkflowExecution(&executionEntity, &executionData)
}

func toWorkflowExecution(
	executionEntity *rdsDbLib.WorkflowExecutionEntity,
	executionData *rdsDbLib.WorkflowExecutionDatum,
) (*structs.WorkflowExecution, error) {
	workflowData := structs.WorkflowEntity{}
	err := structs.UnmarshalOmitEmpty(executionData.WorkflowData, &workflowData)
	if err != nil {
		return nil, err
	}

	// Handle old execution data:
	// old execution data is flatten json, new execution data is struct json
	runExecutionData := structs.WorkflowRunExecutionData{}
	var arrayObject []interface{}
	err = json.Unmarshal([]byte(executionData.Data), &arrayObject)
	if err == nil {
		// Old execution data need unflatten
		unflattenResult, err := UnflattenString(executionData.Data)
		if err != nil {
			return nil, err
		}
		err = UnmarshalOmitEmpty([]byte(unflattenResult), &runExecutionData)
		if err != nil {
			return nil, err
		}
	} else {
		// New execution data can unmarshal directly
		err := UnmarshalOmitEmpty([]byte(executionData.Data), &runExecutionData)
		if err != nil {
			return nil, err
		}
	}

	return &structs.WorkflowExecution{
		Id:             fmt.Sprint(executionEntity.ID),
		Data:           &runExecutionData,
		Finished:       executionEntity.Finished,
		Mode:           structs.WorkflowExecutionMode(executionEntity.Mode),
		Status:         structs.WorkflowExecutionStatus(executionEntity.Status.String),
		RetryOf:        executionEntity.RetryOf.String,
		RetrySuccessId: executionEntity.RetrySuccessId.String,
		StartedAt:      &executionEntity.StartedAt,
		StoppedAt:      ConvertNullTimeToStandardTimePointer(executionEntity.StoppedAt),
		WaitTill:       ConvertNullTimeToStandardTimePointer(executionEntity.WaitTill),
		WorkflowData:   &workflowData,
		WorkflowId:     executionEntity.WorkflowId,
	}, nil
}

func makeNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	} else {
		return sql.NullTime{Time: *t, Valid: true}
	}
}

// Update the workflow execution entity and data.
func UpdateWorkflowExecutionEntityAndData(
	ctx context.Context, executionId int32, workflowExecutionEntity *structs.WorkflowExecution) error {
	_, err := rdsDbQueries.UpdateWorkflowExecutionEntity(
		ctx,
		rdsDbLib.UpdateWorkflowExecutionEntityParams{
			ID:             executionId,
			Finished:       workflowExecutionEntity.Finished,
			Mode:           string(workflowExecutionEntity.Mode),
			RetryOf:        sql.NullString{String: workflowExecutionEntity.RetryOf, Valid: true},
			RetrySuccessId: sql.NullString{String: workflowExecutionEntity.RetrySuccessId, Valid: true},
			StoppedAt:      makeNullTime(workflowExecutionEntity.StoppedAt),
			WaitTill:       makeNullTime(workflowExecutionEntity.WaitTill),
			Status:         sql.NullString{String: string(workflowExecutionEntity.Status), Valid: true},
		})
	if err != nil {
		return err
	}
	executionDataParams := rdsDbLib.UpdateWorkflowExecutionDataParams{
		ExecutionId: executionId,
	}
	if workflowExecutionEntity.Data != nil || workflowExecutionEntity.WorkflowData != nil {
		if workflowExecutionEntity.Data != nil {
			dataBytes, err := json.Marshal(workflowExecutionEntity.Data)
			if err != nil {
				return err
			}
			executionDataParams.Data = string(dataBytes)
		}
		if workflowExecutionEntity.WorkflowData != nil {
			workflowDataBytes, err := json.Marshal(workflowExecutionEntity.WorkflowData)
			if err != nil {
				return err
			}
			executionDataParams.WorkflowData = workflowDataBytes
		}
		_, err = rdsDbQueries.UpdateWorkflowExecutionData(ctx, executionDataParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateWorkflowExecutionAndData(ctx context.Context,
	data *structs.ExecutingWorkflowData) (*rdsDbLib.WorkflowExecutionEntity, error) {
	entity, err := rdsDbQueries.CreateWorkflowExecutionEntity(
		ctx,
		rdsDbLib.CreateWorkflowExecutionEntityParams{
			Finished:       false,
			Mode:           string(data.ExecutionData.ExecutionMode),
			RetryOf:        sql.NullString{},
			RetrySuccessId: sql.NullString{},
			StartedAt:      *data.StartedAt,
			Status:         sql.NullString{String: string(data.Status), Valid: true},
			WorkflowId:     data.ExecutionData.WorkflowData.ID,
		})
	if err != nil {
		return nil, err
	}

	workflowDataJson, err := json.Marshal(data.ExecutionData.WorkflowData)
	if err != nil {
		return nil, err
	}

	workflowRunExecutionDataStr := "{}"
	if data.ExecutionData.ExecutionData != nil {
		workflowRunExecutionDataJson, err := json.Marshal(data.ExecutionData.ExecutionData)
		if err != nil {
			return nil, err
		}
		workflowRunExecutionDataStr = string(workflowRunExecutionDataJson)
	}

	_, err = rdsDbQueries.CreateWorkflowExecutionData(
		ctx,
		rdsDbLib.CreateWorkflowExecutionDataParams{
			ExecutionId:  entity.ID,
			WorkflowData: workflowDataJson,
			Data:         workflowRunExecutionDataStr,
		})
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

// Delete both the workflow execution entity and data.
func DeleteWorkflowExecutionAndData(ctx context.Context, workflowId string, executionId int32) error {
	err := rdsDbQueries.DeleteWorkflowExecutionEntity(ctx, rdsDbLib.DeleteWorkflowExecutionEntityParams{
		WorkflowId: workflowId,
		ID:         executionId,
	})
	if err != nil {
		return err
	}
	err = rdsDbQueries.DeleteWorkflowExecutionData(ctx, rdsDbLib.DeleteWorkflowExecutionDataParams{
		WorkflowID:  workflowId,
		ExecutionId: executionId,
	})
	return err
}
