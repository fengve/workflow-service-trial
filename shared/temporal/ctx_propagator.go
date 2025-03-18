package temporal

import (
	"context"

	sharedlog "github.com/sugerio/workflow-service-trial/shared/log"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type (
	// propagator implements the context propagator used by common
	commonPropagator struct{}

	CommonCtxPropagation struct {
		Environment *structs.Environment `json:"environment,omitempty"`
		TraceId     string               `json:"traceId,omitempty"`
	}
)

// PropagateKey is the key used to store the common value in the Context object
const CommonPropagateContextKey = "common_propagate_ctx_key"

// propagationKey is the key used by the propagator to pass values through the
// Temporal server headers
const CommonPropagateHeaderKey = "common_propagate_header_key"

func SetContext(ctx context.Context, env *structs.Environment) context.Context {
	traceId, _ := ctx.Value("TraceId").(string)
	ctx = context.WithValue(
		ctx,
		CommonPropagateContextKey,
		CommonCtxPropagation{Environment: env, TraceId: traceId})
	return ctx
}

// NewCommonContextPropagator returns a common context propagator that propagates a set of
// string key-value pairs across a workflow
func NewCommonContextPropagator() workflow.ContextPropagator {
	return &commonPropagator{}
}

// Inject injects values from context into headers for propagation
func (s *commonPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(CommonPropagateContextKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(CommonPropagateHeaderKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *commonPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(CommonPropagateContextKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(CommonPropagateHeaderKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *commonPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(CommonPropagateHeaderKey); ok {
		var commonCtxPropagation CommonCtxPropagation
		if err := converter.GetDefaultDataConverter().FromPayload(value, &commonCtxPropagation); err != nil {
			return ctx, nil
		}
		ctx = context.WithValue(ctx, CommonPropagateContextKey, commonCtxPropagation)
		ctx = context.WithValue(ctx, "TraceId", commonCtxPropagation.TraceId)
		logger := sharedlog.GetLogger(ctx)
		ctx = sharedlog.SetLogger(ctx, logger)
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *commonPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(CommonPropagateHeaderKey); ok {
		var commonCtxPropagation CommonCtxPropagation
		if err := converter.GetDefaultDataConverter().FromPayload(value, &commonCtxPropagation); err != nil {
			return ctx, nil
		}
		ctx = workflow.WithValue(ctx, CommonPropagateContextKey, commonCtxPropagation)
		ctx = workflow.WithValue(ctx, "TraceId", commonCtxPropagation.TraceId)
		logger := sharedlog.GetLogger(ctx)
		ctx = sharedlog.SetWorkflowLogger(ctx, logger)
	}
	return ctx, nil
}
