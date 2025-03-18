package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/sugerio/workflow-service-trial/shared/structs"
)

var (
	nodeObjectRegistry = map[string]NodeObject{}
	nodeEmbedIcons     = map[string][]byte{}
)

// NodeObject is the interface for all nodes.
type NodeObject interface {
	// Category returns the node category of itself.
	Category() structs.NodeObjectCategory

	// Name returns the name of the node.
	Name() string

	// DefaultSpec returns the default spec.
	// It must return a pointer to point to a struct.
	DefaultSpec() interface{}

	Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult
}

// NodeMethods is used for dynamic parameters api
// The Methods() must be implemented in the target node.go
type NodeMethods interface {
	Methods() *structs.NodeMethods
}

// webhookMethods in n8n, only handle the default named methods (there is no setup named methods now).
type NodeWebhookMethods interface {
	WebhookMethods() *structs.NodeWebhookMethods
}

type TriggerObject interface {
	Trigger(ctx context.Context, input *structs.WorkflowNode) string
}

// Register registers object.
func Register(o NodeObject) {
	if o.Category() == "" {
		panic(fmt.Errorf("%T: empty kind", o))
	}

	existedObject, existed := nodeObjectRegistry[o.Name()]
	if existed {
		panic(fmt.Errorf("%T and %T got same kind: %s", o, existedObject, o.Name()))
	}

	// Checking object type.
	nodeObjectType := reflect.TypeOf(o)
	if nodeObjectType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%s: want a pointer, got %s", o.Name(), nodeObjectType.Kind()))
	}
	if nodeObjectType.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("%s elem: want a struct, got %s", o.Name(), nodeObjectType.Kind()))
	}

	// Checking spec type.
	specType := reflect.TypeOf(o.DefaultSpec())
	if specType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%s spec: want a pointer, got %s", o.Name(), specType.Kind()))
	}
	if specType.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("%s spec elem: want a struct, got %s", o.Name(), specType.Elem().Kind()))
	}

	nodeObjectRegistry[o.Name()] = o
}

// RegisterEmbedIcons call this if there is embed icon after calling Register
func RegisterEmbedIcons(name string, icon []byte) {
	nodeEmbedIcons[name] = icon
}

func GetAllNodeObjects() map[string]NodeObject {
	return nodeObjectRegistry
}

func GetAllNodeEmbedIcons() map[string][]byte {
	return nodeEmbedIcons
}
