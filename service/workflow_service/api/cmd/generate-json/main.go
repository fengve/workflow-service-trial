package main

import (
	"encoding/json"
	"os"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"

	_ "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes"
)

func main() {
	allNodes := core.GetAllNodeObjects()

	_, err := os.Stat("statics")
	if os.IsNotExist(err) {
		err = os.Mkdir("statics", 0755)
		if err != nil {
			panic(err)
		}
	}

	// write to json file in workflow_service/api/statics/nodes.json
	fd, err := os.Create("statics/nodes.json")
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	var resp []interface{}
	for _, nodes := range allNodes {
		spec, ok := nodes.DefaultSpec().(*structs.WorkflowNodeSpec)
		if !ok {
			panic("invalid spec")
		}
		var data interface{}
		_ = json.Unmarshal(spec.JsonConfig, &data)
		resp = append(resp, data)
	}

	raw, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		panic(err)
	}
	_, _ = fd.Write(raw)
}
