// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import "github.com/larksuite/cli/shortcuts/common"

type workflowWriteOperation uint8

const (
	workflowWriteCreate workflowWriteOperation = iota
	workflowWriteUpdate
)

func parseWorkflowBodyJSON(runtime *common.RuntimeContext, operation workflowWriteOperation) (map[string]interface{}, error) {
	pc := newParseCtx(runtime)
	body, err := parseJSONObject(pc, runtime.Str("json"), "json")
	if err != nil {
		return nil, err
	}
	if err := validateWorkflowBodyForCLI(body, operation); err != nil {
		return nil, err
	}
	return body, nil
}

func validateWorkflowBodyForCLI(body map[string]interface{}, operation workflowWriteOperation) error {
	steps, ok := body["steps"].([]interface{})
	if !ok {
		return nil
	}

	var stepIDs map[string]int
	for stepIndex, rawStep := range steps {
		step, ok := rawStep.(map[string]interface{})
		if !ok {
			continue
		}

		stepType, _ := step["type"].(string)
		switch stepType {
		case aiAnalysisActionType:
			if err := validateWorkflowAIAnalysisAction(stepIndex, step); err != nil {
				return err
			}
		case workflowAIClassificationStepType:
			if stepIDs == nil {
				stepIDs = indexWorkflowStepIDs(steps)
			}
			if err := validateWorkflowAIClassificationStep(stepIndex, step, stepIDs, operation); err != nil {
				return err
			}
		}
	}

	return nil
}
